# Copyright 2017 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""This package implements the Python package layer builder."""

import logging
import os
import subprocess
import tempfile

from ftl.common import constants
from ftl.common import ftl_util
from ftl.common import single_layer_image
from ftl.common import tar_to_dockerimage


class PackageLayerBuilder(single_layer_image.CacheableLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 pkg_dir=None,
                 dep_img_lyr=None,
                 cache=None):
        super(PackageLayerBuilder, self).__init__()
        self._ctx = ctx
        self._pkg_dir = pkg_dir
        self._descriptor_files = descriptor_files
        self._dep_img_lyr = dep_img_lyr
        self._cache = cache

    def GetCacheKeyRaw(self):
        descriptor_contents = ftl_util.descriptor_parser(
            self._descriptor_files, self._ctx)
        return '%s %s' % (descriptor_contents,
                          self._dep_img_lyr.GetCacheKeyRaw())

    def BuildLayer(self):
        cached_img = None
        if self._cache:
            with ftl_util.Timing('Checking cached pkg layer'):
                key = self.GetCacheKey()
                cached_img = self._cache.Get(key)
                self._log_cache_result(False if cached_img is None else True)
        if cached_img:
            self.SetImage(cached_img)
        else:
            with ftl_util.Timing('Building pkg layer'):
                self._build_layer()
            if self._cache:
                with ftl_util.Timing('Uploading pkg layer'):
                    self._cache.Set(self.GetCacheKey(), self.GetImage())

    def _build_layer(self):
        blob, u_blob = ftl_util.zip_dir_to_layer_sha(self._pkg_dir)
        overrides = ftl_util.generate_overrides(False)
        self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob], overrides)

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_1_CACHE_HIT
        else:
            cache_str = constants.PHASE_1_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PYTHON (package)',
                key=self.GetCacheKey()))


class RequirementsLayerBuilder(single_layer_image.CacheableLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 pkg_dir=None,
                 dep_img_lyr=None,
                 wheel_dir=None,
                 venv_dir=None,
                 pip_cmd=None,
                 venv_cmd=None,
                 cache=None):
        super(RequirementsLayerBuilder, self).__init__()
        self._ctx = ctx
        self._pkg_dir = pkg_dir
        self._wheel_dir = wheel_dir
        self._venv_dir = venv_dir
        self._pip_cmd = pip_cmd
        self._venv_cmd = venv_cmd
        self._descriptor_files = descriptor_files
        self._dep_img_lyr = dep_img_lyr
        self._cache = cache

    def GetCacheKeyRaw(self):
        descriptor_contents = ftl_util.descriptor_parser(
            self._descriptor_files, self._ctx)
        return '%s %s' % (descriptor_contents,
                          self._dep_img_lyr.GetCacheKeyRaw())

    def BuildLayer(self):
        cached_img = None
        if self._cache:
            with ftl_util.Timing('Checking cached req.txt layer'):
                key = self.GetCacheKey()
                cached_img = self._cache.Get(key)
                self._log_cache_result(False if cached_img is None else True)
        if cached_img:
            self.SetImage(cached_img)
        else:
            with ftl_util.Timing('Installing pip packages'):
                pkg_descriptor = ftl_util.descriptor_parser(
                    self._descriptor_files, self._ctx)
                self._pip_install(pkg_descriptor)

            with ftl_util.Timing('Resolving whl paths'):
                whls = self._resolve_whls()
                pkg_dirs = [self._whl_to_fslayer(whl) for whl in whls]

            req_txt_imgs = []
            for whl_pkg_dir in pkg_dirs:
                layer_builder = PackageLayerBuilder(
                    ctx=self._ctx,
                    descriptor_files=self._descriptor_files,
                    pkg_dir=whl_pkg_dir,
                    dep_img_lyr=self,
                    cache=self._cache)
                layer_builder.BuildLayer()
                req_txt_imgs.append(layer_builder.GetImage())

            req_txt_image = ftl_util.AppendLayersIntoImage(req_txt_imgs)

            self.SetImage(req_txt_image)

            if self._cache:
                with ftl_util.Timing('Uploading req.txt image'):
                    self._cache.Set(self.GetCacheKey(), self.GetImage())

    def _resolve_whls(self):
        return [
            os.path.join(self._wheel_dir, f)
            for f in os.listdir(self._wheel_dir)
        ]

    def _whl_to_fslayer(self, whl):
        tmp_dir = tempfile.mkdtemp()
        pkg_dir = os.path.join(tmp_dir, 'env')
        os.makedirs(pkg_dir)

        pip_cmd_args = list(self._pip_cmd)
        pip_cmd_args.extend(['install', '--no-deps', '--prefix', pkg_dir, whl])

        proc_pipe = subprocess.Popen(
            pip_cmd_args,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=self._gen_pip_env(),
        )
        stdout, stderr = proc_pipe.communicate()
        logging.info("`pip install` stdout:\n%s" % stdout)
        if stderr:
            logging.error("`pip install` had error output:\n%s" % stderr)
        if proc_pipe.returncode:
            raise Exception("error: `pip install` returned code: %d" %
                            proc_pipe.returncode)
        return tmp_dir

    def _pip_install(self, pkg_txt):
        with ftl_util.Timing('pip_download_wheels'):
            pip_cmd_args = list(self._pip_cmd)
            pip_cmd_args.extend(
                ['wheel', '-w', self._wheel_dir, '-r', '/dev/stdin'])

            proc_pipe = subprocess.Popen(
                pip_cmd_args,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=self._gen_pip_env(),
            )
            stdout, stderr = proc_pipe.communicate(input=pkg_txt)
            logging.info("`pip wheel` stdout:\n%s" % stdout)
            if stderr:
                logging.error("`pip wheel` had error output:\n%s" % stderr)
            if proc_pipe.returncode:
                raise Exception("error: `pip wheel` returned code: %d" %
                                proc_pipe.returncode)

    def _gen_pip_env(self):
        pip_env = os.environ.copy()
        # bazel adds its own PYTHONPATH to the env
        # which must be removed for the pip calls to work properly
        pip_env.pop('PYTHONPATH', None)
        pip_env['VIRTUAL_ENV'] = self._venv_dir
        pip_env['PATH'] = self._venv_dir + '/bin' + ':' + os.environ['PATH']
        return pip_env

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_1_CACHE_HIT
        else:
            cache_str = constants.PHASE_1_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PYTHON (requirements)',
                key=self.GetCacheKey()))


class PipfileLayerBuilder(RequirementsLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 pkg_descriptor=None,
                 pkg_dir=None,
                 dep_img_lyr=None,
                 wheel_dir=None,
                 venv_dir=None,
                 pip_cmd=None,
                 venv_cmd=None,
                 cache=None):
        super(RequirementsLayerBuilder, self).__init__()
        self._ctx = ctx
        self._pkg_dir = pkg_dir
        self._wheel_dir = wheel_dir
        self._venv_dir = venv_dir
        self._pip_cmd = pip_cmd
        self._venv_cmd = venv_cmd
        self._descriptor_files = descriptor_files
        self._dep_img_lyr = dep_img_lyr
        self._cache = cache
        self._pkg_descriptor = pkg_descriptor

    def GetCacheKeyRaw(self):
        return "%s %s %s" % (self._pkg_descriptor[0], self._pkg_descriptor[1],
                             self._dep_img_lyr.GetCacheKeyRaw())

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_2_CACHE_HIT
        else:
            cache_str = constants.PHASE_2_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PYTHON',
                package_name=self._pkg_descriptor[0],
                package_version=self._pkg_descriptor[1],
                key=self.GetCacheKey()))

    def BuildLayer(self):
        cached_img = None
        if self._cache:
            with ftl_util.Timing('Checking cached pkg layer'):
                key = self.GetCacheKey()
                cached_img = self._cache.Get(key)
                self._log_cache_result(False if cached_img is None else True)
        if cached_img:
            self.SetImage(cached_img)
        else:
            with ftl_util.Timing('Installing pip packages'):
                self._pip_install(' '.join(self._pkg_descriptor))

            with ftl_util.Timing('Resolving whl paths'):
                whls = self._resolve_whls()
                if len(whls) != 1:
                    raise Exception("expected one whl for one installed pkg")
                pkg_dir = self._whl_to_fslayer(whls[0])
                blob, u_blob = ftl_util.zip_dir_to_layer_sha(pkg_dir)
                overrides = ftl_util.generate_overrides(False)
                self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob],
                                                           overrides)

    def _pip_install(self, pkg_txt):
        with ftl_util.Timing('pip_download_wheel'):
            pip_cmd_args = list(self._pip_cmd)
            pip_cmd_args.extend(
                ['wheel', '-w', self._wheel_dir, '-r', '/dev/stdin'])
            pip_cmd_args.extend(['--no-deps'])

            proc_pipe = subprocess.Popen(
                pip_cmd_args,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=self._gen_pip_env(),
            )
            stdout, stderr = proc_pipe.communicate(input=pkg_txt)
            logging.info("`pip wheel` stdout:\n%s" % stdout)
            if stderr:
                logging.error("`pip wheel` had error output:\n%s" % stderr)
            if proc_pipe.returncode:
                raise Exception("error: `pip wheel` returned code: %d" %
                                proc_pipe.returncode)


class InterpreterLayerBuilder(single_layer_image.CacheableLayerBuilder):
    def __init__(self,
                 venv_dir=None,
                 python_cmd=None,
                 venv_cmd=None,
                 cache=None):
        super(InterpreterLayerBuilder, self).__init__()
        self._venv_dir = venv_dir
        self._python_cmd = python_cmd
        self._venv_cmd = venv_cmd
        self._cache = cache

    def GetCacheKeyRaw(self):
        return '%s %s' % (self._python_cmd, self._venv_cmd)

    def BuildLayer(self):
        cached_img = None
        if self._cache:
            with ftl_util.Timing('Checking cached pkg layer'):
                key = self.GetCacheKey()
                cached_img = self._cache.Get(key)
                self._log_cache_result(False if cached_img is None else True)
        if cached_img:
            self.SetImage(cached_img)
        else:
            with ftl_util.Timing('Building pkg layer'):
                self._build_layer()
            if self._cache:
                with ftl_util.Timing('Uploading pkg layer'):
                    self._cache.Set(self.GetCacheKey(), self.GetImage())

    def _build_layer(self):
        self._setup_venv()

        tar_path = tempfile.mktemp(suffix='.tar')
        with ftl_util.Timing('tar_runtime_package'):
            subprocess.check_call(['tar', '-cf', tar_path, self._venv_dir])

        u_blob = open(tar_path, 'r').read()
        # We use gzip for performance instead of python's zip.
        with ftl_util.Timing('gzip_runtime_tar'):
            subprocess.check_call(['gzip', tar_path, '-1'])
        blob = open(os.path.join(tar_path + '.gz'), 'rb').read()

        overrides = ftl_util.generate_overrides(True)
        self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob], overrides)

    def _setup_venv(self):
        with ftl_util.Timing('create_virtualenv'):
            venv_cmd_args = list(self._venv_cmd)
            venv_cmd_args.extend([
                '--no-download',
                self._venv_dir,
                '-p',
            ])
            venv_cmd_args.extend(self._python_cmd)
            proc_pipe = subprocess.Popen(
                venv_cmd_args,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            stdout, stderr = proc_pipe.communicate()
            logging.info("`virtualenv` stdout:\n%s" % stdout)
            if stderr:
                logging.error("`virtualenv` had error output:\n%s" % stderr)
            if proc_pipe.returncode:
                raise Exception("error: `virtualenv` returned code: %d" %
                                proc_pipe.returncode)

            subprocess.check_call(venv_cmd_args)

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_1_CACHE_HIT
        else:
            cache_str = constants.PHASE_1_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PYTHON (interpreter)',
                key=self.GetCacheKey()))
