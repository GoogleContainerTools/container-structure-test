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

from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
import unittest
import mock_registry
import mock
import mock_from_registry
import mock_session_push


class MockRegistryTestBase(unittest.TestCase):
    """A base class for tests that need to mock docker_image.FromRegistry
       or docker_session.Push"""

    def startObjectPatch(self, *args, **kwargs):
        patcher = mock.patch.object(*args, **kwargs)
        return patcher.start()

    def setUp(self):
        self.startObjectPatch(docker_image, 'FromRegistry',
                              new=mock_from_registry.MockFromRegistry)
        self.startObjectPatch(docker_session, 'Push',
                              new=mock_session_push.MockSessionPush)
        self.registry = mock_registry.MockRegistry()
        mock_from_registry.MockFromRegistry().setRegistry(self.registry)
        mock_session_push.MockSessionPush().setRegistry(self.registry)

    def AssertPushed(self, registry, name):
        self.assertTrue(registry.existsImage(str(name)))

    def AssertNotPushed(self, registry, name):
        self.assertFalse(registry.existsImage(str(name)))
