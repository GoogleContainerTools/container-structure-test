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
import mock_registry


class MockFromRegistry(docker_image.FromRegistry):

    REGISTRY = mock_registry.MockRegistry()

    def __init__(self, name="", basic_creds=None, transport=None):
        self._name = name

    def setRegistry(self, registry):
        if isinstance(registry, mock_registry.MockRegistry):
            MockFromRegistry.REGISTRY = registry

    def exists(self):
        return MockFromRegistry.REGISTRY.existsImage(self._name)

    def manifests(self):
        return MockFromRegistry.REGISTRY.getManifests(self._name)

    def tags(self):
        return MockFromRegistry.REGISTRY.getTags(self._name)

    def getName(self):
        return self._name

    # __enter__ and __exit__ allow use as a context manager.
    def __enter__(self):
        return self

    def __exit__(self, unused_type, unused_value, unused_traceback):
        pass
