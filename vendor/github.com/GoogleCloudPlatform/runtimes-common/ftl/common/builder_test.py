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

import cStringIO
import os
import unittest
import tarfile

import builder
import layer_builder
import context


class JustAppTest(unittest.TestCase):
    def test_build_app_layer(self):
        # All the files in the context should be added to the layer.

        files = {
            'foo': 'foo_contents',
            'bar': 'bar_contents',
            'baz/bat': 'bat_contents'
        }
        ctx = context.Memory()
        for p, f in files.iteritems():
            ctx.AddFile(p, f)

        b = builder.JustApp(ctx)
        app_builder = layer_builder.AppLayerBuilder(b._ctx)
        app_builder.BuildLayer()
        app_layer = app_builder.GetImage().GetFirstBlob()
        stream = cStringIO.StringIO(app_layer)
        with tarfile.open(fileobj=stream, mode='r:gz') as tf:
            self.assertEqual(len(tf.getnames()), len(files))
            for p, f in files.iteritems():
                tar_path = os.path.join('srv', p)
                self.assertEquals(tf.extractfile(tar_path).read(), f)


if __name__ == '__main__':
    unittest.main()
