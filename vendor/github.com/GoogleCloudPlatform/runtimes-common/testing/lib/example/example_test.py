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

from testing.lib import mock_registry_test_base
from containerregistry.client.v2_2 import docker_image
import example
import unittest

DIGEST = '0000000000000000000000000000000000000000000000000000000000000000'


class ExampleTest(mock_registry_test_base.MockRegistryTestBase):

    def setUp(self):
        super(ExampleTest, self).setUp()

    def testMain(self):
        # Add initial image to registry
        with docker_image.FromRegistry(
                'fake.gcr.io/test/test@sha256:' + DIGEST) as img:
            self.registry.setImage('fake.gcr.io/test/test@sha256:' + DIGEST,
                                   img)

        example.main()
        # Assert that new image was pushed correctly
        self.AssertPushed(self.registry, "fake.gcr.io/test/test:tag")


if __name__ == '__main__':
    unittest.main()
