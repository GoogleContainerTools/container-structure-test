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

from ftl.cached import cached
import unittest
import mock


class cachedTest(unittest.TestCase):
    def setUp(self):
        _RUNTIME = "test"
        args = mock.Mock()
        args.name = 'gcr.io/test/test:latest'
        args.base = 'gcr.io/google-appengine/python:latest'
        args.directory = '/test'
        self.c = cached.Cached(args, _RUNTIME)

    def testCompareLayersNoError(self):
        lyr_shas_1 = set(["a", "b", "c", "d"])
        lyr_shas_2 = set(["a", "b", "c", "e"])
        try:
            self.c._compare_layers(lyr_shas_1, lyr_shas_2)
        except RuntimeError:
            self.fail("_compare_layers raised RuntimeError unexpectedly!")

    def testCompareLayersError(self):
        lyr_shas_1 = set(["a", "b", "c", "d"])
        lyr_shas_2 = set(["a", "b", "e", "f`"])
        with self.assertRaises(RuntimeError):
            self.c._compare_layers(lyr_shas_1, lyr_shas_2)


if __name__ == '__main__':
    unittest.main()
