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
"""Unit tests for cache.py"""

import unittest

import cache

import mock


class RegistryTest(unittest.TestCase):
    def setUp(self):
        pass

    @mock.patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_get(self, mock_from):
        # The cache needs to calculate a digest from the base image.
        fake_base = mock.MagicMock()
        fake_base.digest.return_value = 'abc123'

        mock_img = mock.MagicMock()
        mock_from.return_value.__enter__.return_value = mock_img

        # Test when the image exists.
        mock_img.exists.return_value = True
        c = cache.Registry(
            repo='fake.gcr.io/google-appengine',
            namespace='namespace',
            creds=None,
            transport=None)
        self.assertEquals(c._getEntry('abc123'), mock_img)

        # Test when it does not exist
        mock_img.exists.return_value = False
        c = cache.Registry(
            repo='fake.gcr.io/google-appengine',
            namespace='namespace',
            creds=None,
            transport=None)
        self.assertIsNone(c._getEntry('abc123'))


if __name__ == '__main__':
    unittest.main()
