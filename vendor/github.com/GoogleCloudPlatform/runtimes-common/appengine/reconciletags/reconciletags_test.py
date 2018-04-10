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

"""Unit tests for reconcile-tags.py.

Unit tests for reconcile-tags.py.
"""
import unittest
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
import mock
from mock import patch
import tag_reconciler

_REGISTRY = 'gcr.io'
_REPO = 'foobar/baz'
_FULL_REPO = _REGISTRY + '/' + _REPO
_DIGEST1 = '0000000000000000000000000000000000000000000000000000000000000000'
_DIGEST2 = '0000000000000000000000000000000000000000000000000000000000000001'
_TAG1 = 'tag1'
_TAG2 = 'tag2'

_LIST_RESP = """
[
  {
    "digest":
        "0000000000000000000000000000000000000000000000000000000000000000",
    "tags": [
      "tag1"
    ],
    "timestamp": {
    }
  }
]
"""


class ReconcileTagsTest(unittest.TestCase):

    def setUp(self):
        self.r = tag_reconciler.TagReconciler()
        self.data = {'projects':
                     [{'base_registry': 'gcr.io',
                       'additional_registries': [],
                       'repository': _REPO,
                       'images': [{'digest': _DIGEST1, 'tag': _TAG1}]}]}

    @patch('tag_reconciler.TagReconciler.get_digest_from_prefix')
    @patch('containerregistry.client.v2_2.docker_session.Push')
    @patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_reconcile_tags(self, mock_from_registry, mock_push,
                            mock_get_digest):
        fake_base = mock.MagicMock()
        fake_base.tags.return_value = [_TAG1]

        mock_img = mock.MagicMock()
        mock_img.__enter__.return_value = fake_base
        mock_from_registry.return_value = mock_img
        mock_push.return_value = docker_session.Push()

        mock_get_digest.return_value = _DIGEST1

        self.r.reconcile_tags(self.data, False)

        assert mock_from_registry.called
        assert mock_push.called

    @patch('tag_reconciler.TagReconciler.get_digest_from_prefix')
    @patch('containerregistry.client.v2_2.docker_session.Push')
    @patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_dry_run(self, mock_from_registry, mock_push, mock_get_digest):
        mock_from_registry.return_value = docker_image.FromRegistry()
        mock_push.return_value = docker_session.Push()
        mock_get_digest.return_value = _DIGEST1

        self.r.reconcile_tags(self.data, True)

        assert mock_from_registry.called
        assert mock_push.called

    @patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_get_existing_tags(self, mock_from_registry):

        fake_base = mock.MagicMock()
        fake_base.tags.return_value = [_TAG1]

        mock_img = mock.MagicMock()
        mock_img.__enter__.return_value = fake_base
        mock_from_registry.return_value = mock_img

        existing_tags = self.r.get_existing_tags(_FULL_REPO, _DIGEST1)

        assert mock_from_registry.called
        self.assertEqual([_TAG1], existing_tags)

    @patch('containerregistry.client.v2_2.docker_session.Push')
    @patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_add_tag(self, mock_from_registry, mock_push):
        mock_from_registry.return_value = docker_image.FromRegistry()
        mock_push.return_value = docker_session.Push()

        self.r.add_tags(_FULL_REPO+'@sha256:'+_DIGEST2,
                        _FULL_REPO+':'+_TAG2, False)

        assert mock_from_registry.called
        assert mock_push.called

    @patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def test_get_digest(self, mock_from_registry):
        fake_base = mock.MagicMock()
        fake_base.manifests.return_value = ["sha256:" + _DIGEST1,
                                            "sha256" + _DIGEST2]

        mock_img = mock.MagicMock()
        mock_img.__enter__.return_value = fake_base
        mock_from_registry.return_value = mock_img

        with self.assertRaises(AssertionError):
            self.r.get_digest_from_prefix(_FULL_REPO, _DIGEST1[0:3])

        digest = self.r.get_digest_from_prefix(_FULL_REPO, _DIGEST1)
        self.assertEqual(digest, _DIGEST1)


if __name__ == '__main__':
    unittest.main()
