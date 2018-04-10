# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import subprocess
import os
import logging
import httplib2
import json

from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
from containerregistry.transport import transport_pool

_THREADS = 32


class Cached():
    def __init__(self, args, runtime):
        self._base = args.base
        self._name = args.name
        self._directory = args.directory
        self._labels = [args.label_1, args.label_2]
        self._runtime = runtime
        logging.getLogger().setLevel("NOTSET")
        logging.basicConfig(
            format='%(asctime)s.%(msecs)03d %(levelname)-8s %(message)s',
            datefmt='%Y-%m-%d,%H:%M:%S')

    def run_cached_tests(self):
        logging.info('Beginning building {0} images'.format(self._runtime))
        # For the binary
        builder_path = 'ftl/{0}_builder.par'.format(self._runtime)

        # For container builder
        if not os.path.isfile(builder_path):
            builder_path = 'bazel-bin/ftl/{0}_builder.par'.format(
                self._runtime)
        lyr_shas = []
        for label in self._labels:
            img_name = ''.join([self._name.split(":")[0], ":", label])
            cmd = subprocess.Popen(
                [
                    builder_path, '--base', self._base, '--name', img_name,
                    '--directory', self._directory
                ],
                stderr=subprocess.PIPE)
            _, output = cmd.communicate()
            logging.info('output of build {0}: {1}'.format(label, output))
            lyr_shas.append(self._fetch_lyr_shas(img_name))
            self._del_img_from_gcr(img_name)
        try:
            self._compare_layers(lyr_shas[0], lyr_shas[1])
        except RuntimeError as e:
            logging.error(e)
            exit(1)

    def _fetch_lyr_shas(self, img_name):
        name = docker_name.Tag(img_name)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http, size=_THREADS)
        with docker_image.FromRegistry(name, creds, transport) as img:
            lyrs = json.loads(img.manifest())['layers']
            lyr_shas = []
            for lyr in lyrs:
                lyr_shas.append(lyr['digest'])
            return set(lyr_shas)

    def _compare_layers(self, lyr_shas_1, lyr_shas_2):
        if len(lyr_shas_1) != len(lyr_shas_2):
            raise RuntimeError("different amount of layers found when \
                               reuploading same image")
        lyr_diff_cnt = 0
        for lyr_sha in lyr_shas_1:
            if lyr_sha not in lyr_shas_2:
                lyr_diff_cnt += 1
            if lyr_diff_cnt > 1:
                raise RuntimeError("more layers differed then app layer when \
                    reuploading image")

    def _del_img_from_gcr(self, img_name):
        img_tag = docker_name.Tag(img_name)
        creds = docker_creds.DefaultKeychain.Resolve(img_tag)
        transport = transport_pool.Http(httplib2.Http, size=_THREADS)
        with docker_image.FromRegistry(img_tag, creds,
                                       transport) as base_image:
            img_digest = docker_name.Digest(''.join(
                [self._name.split(":")[0], "@",
                 str(base_image.digest())]))

        logging.info('Deleting tag {0}'.format(img_tag))
        docker_session.Delete(img_tag, creds, transport)
        logging.info('Deleting image {0}'.format(img_digest))
        docker_session.Delete(img_digest, creds, transport)
        return
