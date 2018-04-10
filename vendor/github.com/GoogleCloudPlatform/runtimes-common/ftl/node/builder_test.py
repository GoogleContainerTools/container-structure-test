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

import json
import unittest
import tempfile
import mock

from ftl.common import context

from ftl.node import builder
from ftl.node import layer_builder

_PACKAGE_JSON = json.loads("""
{
  "name": "hello-world-express",
  "description": "Hello World test app",
  "version": "0.0.1",
  "private": true,
  "dependencies": {
    "express": "3.x"
  }
}
""")
_PACKAGE_JSON_TEXT = json.dumps(_PACKAGE_JSON)

_APP = """
var express = require('express');
var app = express();

// Routes
app.get('/', function(req, res) {
  res.send('Hello World!');
});

// Listen
var port = process.env.PORT || 3000;
app.listen(port);
console.log('Listening on localhost:'+ port);
"""


class NodeTest(unittest.TestCase):
    @mock.patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def setUp(self, mock_from):
        mock_from.return_value.__enter__.return_value = None
        self._tmpdir = tempfile.mkdtemp()
        self.ctx = context.Memory()
        self.ctx.AddFile("app.js", _APP)
        args = mock.Mock()
        args.name = 'gcr.io/test/test:latest'
        args.base = 'gcr.io/google-appengine/python:latest'
        args.tar_base_image_path = None
        self.builder = builder.Node(self.ctx, args)
        self.layer_builder = layer_builder.LayerBuilder(
            ctx=self.builder._ctx,
            descriptor_files=self.builder._descriptor_files,
            destination_path="/app")

        # Mock out the calls to package managers for speed.
        self.layer_builder._gen_npm_install_tar = mock.Mock()
        self.layer_builder._gen_npm_install_tar.return_value = ('layer', 'sha')

    @mock.patch('ftl.common.tar_to_dockerimage.FromFSImage.uncompressed_blob')
    def test_create_package_base_no_descriptor(self, mock_from):
        mock_from.return_value = "layer"
        self.assertFalse(self.ctx.Contains('package.json'))
        self.assertFalse(self.ctx.Contains('package-lock.json'))

        self.layer_builder.BuildLayer()
        self.assertIsInstance(self.layer_builder.GetImage().GetFirstBlob(),
                              str)


if __name__ == '__main__':
    unittest.main()
