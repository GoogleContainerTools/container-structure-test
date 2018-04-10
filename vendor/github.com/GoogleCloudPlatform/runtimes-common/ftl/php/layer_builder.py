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
"""This package implements the PHP package layer builder."""

import logging
import os
import subprocess
import tempfile
import datetime

from ftl.common import constants
from ftl.common import ftl_util
from ftl.common import single_layer_image
from ftl.common import tar_to_dockerimage


class PhaseOneLayerBuilder(single_layer_image.CacheableLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 destination_path=constants.DEFAULT_DESTINATION_PATH,
                 cache=None):
        super(PhaseOneLayerBuilder, self).__init__()
        self._ctx = ctx
        self._descriptor_files = descriptor_files
        self._destination_path = destination_path
        self._cache = cache

    def GetCacheKeyRaw(self):
        return "%s %s" % (
            ftl_util.descriptor_parser(self._descriptor_files, self._ctx),
            self._destination_path)

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
        blob, u_blob = self._gen_composer_install_tar(self._destination_path)
        overrides_dct = {'created': str(datetime.date.today()) + 'T00:00:00Z'}
        self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob],
                                                   overrides_dct)

    def _gen_composer_install_tar(self, destination_path):
        # Create temp directory to write package descriptor to
        pkg_dir = tempfile.mkdtemp()
        app_dir = os.path.join(pkg_dir, destination_path.strip("/"))
        os.makedirs(app_dir)

        # Copy out the relevant package descriptors to a tempdir.
        ftl_util.descriptor_copy(self._ctx, self._descriptor_files, app_dir)

        subprocess.check_call(['rm', '-rf', os.path.join(app_dir, 'vendor')])

        with ftl_util.Timing('Composer_install'):
            subprocess.check_call(
                ['composer', 'install', '--no-dev', '--no-scripts'],
                cwd=app_dir)
        return ftl_util.zip_dir_to_layer_sha(pkg_dir)

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_1_CACHE_HIT
        else:
            cache_str = constants.PHASE_1_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PHP',
                key=self.GetCacheKey()))


class PhaseTwoLayerBuilder(PhaseOneLayerBuilder):
    def __init__(self,
                 ctx=None,
                 descriptor_files=None,
                 destination_path=constants.DEFAULT_DESTINATION_PATH,
                 cache=None,
                 pkg_descriptor=None):
        super(PhaseTwoLayerBuilder, self).__init__()
        self._ctx = ctx
        self._descriptor_files = descriptor_files
        self._pkg_descriptor = pkg_descriptor
        self._destination_path = destination_path
        self._cache = cache

    def GetCacheKeyRaw(self):
        return "%s %s %s" % (self._pkg_descriptor[0], self._pkg_descriptor[1],
                             self._destination_path)

    def _build_layer(self):
        blob, u_blob = self._gen_composer_install_tar(self._destination_path,
                                                      self._pkg_descriptor)
        overrides_dct = {'created': str(datetime.date.today()) + 'T00:00:00Z'}
        self._img = tar_to_dockerimage.FromFSImage([blob], [u_blob],
                                                   overrides_dct)

    def _gen_composer_install_tar(self, destination_path, pkg_descriptor):
        # Create temp directory to write package descriptor to
        pkg_dir = tempfile.mkdtemp()
        app_dir = os.path.join(pkg_dir, destination_path.strip("/"))
        os.makedirs(app_dir)
        subprocess.check_call(['rm', '-rf', os.path.join(app_dir, 'vendor')])

        with ftl_util.Timing('Composer_install'):
            pkg, version = pkg_descriptor
            subprocess.check_call(
                ['composer', 'require',
                 str(pkg), str(version)], cwd=app_dir)
        return ftl_util.zip_dir_to_layer_sha(pkg_dir)

    def _log_cache_result(self, hit):
        if hit:
            cache_str = constants.PHASE_2_CACHE_HIT
        else:
            cache_str = constants.PHASE_2_CACHE_MISS
        logging.info(
            cache_str.format(
                key_version=constants.CACHE_KEY_VERSION,
                language='PHP',
                package_name=self._pkg_descriptor[0],
                package_version=self._pkg_descriptor[1],
                key=self.GetCacheKey()))
