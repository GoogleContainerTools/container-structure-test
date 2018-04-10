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

from containerregistry.client.v2_2 import docker_session
import mock_registry


class MockSessionPush(docker_session.Push):

    REGISTRY = mock_registry.MockRegistry()

    def __init__(self, name="", creds="", transport="", mount=None, threads=1):
        self._name = name
        self._transport = transport
        self._mount = mount
        self._threads = threads

    def setRegistry(self, registry):
        if isinstance(registry, mock_registry.MockRegistry):
            MockSessionPush.REGISTRY = registry

    def upload(self, src_image, use_digest=False):
        if not src_image.exists():
            raise AssertionError("{0} does not exist in registry".format(
                                (str(src_image))))

        dest_img = MockSessionPush.REGISTRY.getImage(src_image)
        MockSessionPush.REGISTRY.setImage(self._name, dest_img)

    # __enter__ and __exit__ allow use as a context manager.
    def __enter__(self):
        return self

    def __exit__(self, exception_type, unused_value, unused_traceback):
        pass
