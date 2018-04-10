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
"""Unit test for context.py"""

import os
import shutil
import tempfile
import unittest

import context


class WorkspaceTest(unittest.TestCase):
    def setUp(self):
        self.tmp_dir = tempfile.mkdtemp()
        self.workspace = context.Workspace(self.tmp_dir)

    def tearDown(self):
        shutil.rmtree(self.tmp_dir)

    def test_contains(self):
        p = 'foo'
        self.assertFalse(self.workspace.Contains(p))
        with open(os.path.join(self.tmp_dir, p), 'w') as f:
            f.write('hey')
        self.assertTrue(self.workspace.Contains(p))

        # Subdir
        d = 'bar'
        p = os.path.join(d, 'baz')

        # Nothing exists yet.
        self.assertFalse(self.workspace.Contains(d))
        self.assertFalse(self.workspace.Contains(p))

        os.makedirs(os.path.join(self.tmp_dir, d), 0777)
        # Contains should still return false for a directory.
        self.assertFalse(self.workspace.Contains(d))

        with open(os.path.join(self.tmp_dir, p), 'w') as f:
            f.write('hey')
        self.assertTrue(self.workspace.Contains(p))


if __name__ == '__main__':
    unittest.main()
