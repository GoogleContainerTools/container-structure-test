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
"""This package implements the Node package layer builder."""

import logging
import os
import subprocess
import tempfile
import json
import datetime

from ftl.common import constants
from ftl.common import ftl_util
from ftl.common import single_layer_image
from ftl.common import tar_to_dockerimage


class LayerBuilder(single_layer_image.CacheableLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 pkg_descriptor=None,
                 destination_path=constants.DEFAULT_DESTINATION_PATH,
                 cache=None):
        super(LayerBuilder, self).__init__()
        self._ctx = ctx
        self._descriptor_files = descriptor_files
        self._pkg_descriptor = pkg_descriptor
        self._destination_path = destination_path
        self._cache = cache

    def GetCacheKeyRaw(self):
        descriptor_contents = ftl_util.descriptor_parser(
            self._descriptor_files, self._ctx)
        return '%s %s' % (descriptor_contents, self._destination_path)

    def BuildLayer(self):
        """Override."""
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
        blob, u_blob = self._gen_npm_install_tar(self._pkg_descriptor,
                                                 self._destination_path)
        self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob],
                                                   self._generate_overrides())

    def _gen_npm_install_tar(self, pkg_descriptor, app_dir):
        try:
            os.makedirs(app_dir)
        except OSError:
            logging.info("%s already exists, skipping creation", app_dir)

        # Copy out the relevant package descriptors to a tempdir.
        ftl_util.descriptor_copy(self._ctx, self._descriptor_files, app_dir)

        self._check_gcp_build(
            json.loads(self._ctx.GetFile(constants.PACKAGE_JSON)), app_dir)
        subprocess.check_call(
            ['rm', '-rf', os.path.join(app_dir, 'node_modules')])
        with ftl_util.Timing("npm_install"):
            if not pkg_descriptor:
                subprocess.check_call(
                    ['npm', 'install', '--production'], cwd=app_dir)
            else:
                subprocess.check_call(
                    ['npm', 'install', '--production', pkg_descriptor],
                    cwd=app_dir)

        tar_path = tempfile.mktemp(suffix='.tar')
        with ftl_util.Timing('tar_runtime_package'):
            subprocess.check_call(['tar', '-cf', tar_path, app_dir])

        u_blob = open(tar_path, 'r').read()
        # We use gzip for performance instead of python's zip.
        with ftl_util.Timing('gzip_runtime_tar'):
            subprocess.check_call(['gzip', tar_path, '-1'])
        return open(os.path.join(tar_path + '.gz'), 'rb').read(), u_blob

    def _generate_overrides(self):
        overrides_dct = {
            'created': str(datetime.date.today()) + "T00:00:00Z",
        }
        return overrides_dct

    def _check_gcp_build(self, package_json, app_dir):
        scripts = package_json.get('scripts', {})
        gcp_build = scripts.get('gcp-build')

        if not gcp_build:
            return

        env = os.environ.copy()
        env["NODE_ENV"] = "development"
        subprocess.check_call(['npm', 'install'], cwd=app_dir, env=env)
        subprocess.check_call(
            ['npm', 'run-script', 'gcp-build'], cwd=app_dir, env=env)

    def _log_cache_result(self, hit):
        if self._pkg_descriptor:
            if hit:
                cache_str = constants.PHASE_2_CACHE_HIT
            else:
                cache_str = constants.PHASE_2_CACHE_MISS
            logging.info(
                cache_str.format(
                    key_version=constants.CACHE_KEY_VERSION,
                    language='NODE',
                    package_name=self._pkg_descriptor[0],
                    package_version=self._pkg_descriptor[1],
                    key=self.GetCacheKey()))
        else:
            if hit:
                cache_str = constants.PHASE_1_CACHE_HIT
            else:
                cache_str = constants.PHASE_1_CACHE_MISS
            logging.info(
                cache_str.format(
                    key_version=constants.CACHE_KEY_VERSION,
                    language='NODE',
                    key=self.GetCacheKey()))
