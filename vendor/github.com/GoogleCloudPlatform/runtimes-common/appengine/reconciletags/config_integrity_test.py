# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Tests to check the integrity of json config files.

These tests assume that the json configs live in a top-level
folder named config."""

import glob
import json
import logging
import os
import unittest
from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.client.v2_2 import docker_image
from containerregistry.transport import transport_pool
from containerregistry.tools import patched
import httplib2


class ReconcilePresubmitTest(unittest.TestCase):

    def _get_digests(self, repo):
        name = docker_name.Repository(repo)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(name, creds, transport) as img:
            digests = [d[len('sha256:'):] for d in img.manifests()]
            return digests
        raise AssertionError('Unable to get digests from {0}'.format(repo))

    def test_json_structure(self):
        for f in glob.glob('../config/tag/*.json'):
            logging.debug('Testing {0}'.format(f))
            with open(f) as tag_map:
                data = json.load(tag_map)
                for project in data['projects']:
                    self.assertEquals(project['base_registry'], 'gcr.io')
                    for registry in project.get('additional_registries', []):
                        self.assertRegexpMatches(registry, '^.*gcr.io$')
                    self.assertIsNotNone(project['repository'])
                    for image in project['images']:
                        self.assertIsInstance(image, dict)
                        self.assertIsNotNone(image['digest'])
                        self.assertIsNotNone(image['tag'])

    def test_digests_are_real(self):
        for f in glob.glob('../config/tag/*.json'):
            logging.debug('Testing {0}'.format(f))
            with open(f) as tag_map:
                data = json.load(tag_map)
                for project in data['projects']:
                    default_registry = project['base_registry']
                    full_repo = os.path.join(default_registry,
                                             project['repository'])
                    logging.debug('Checking {0}'.format(full_repo))
                    digests = self._get_digests(full_repo)
                    for image in project['images']:
                        logging.debug('Checking {0}'
                                      .format(image['digest']))
                        self.assertTrue(any(
                                        digest.startswith(image['digest'])
                                        for digest in digests))


if __name__ == '__main__':
    with patched.Httplib2():
        logging.basicConfig(level=logging.DEBUG)
        unittest.main()
