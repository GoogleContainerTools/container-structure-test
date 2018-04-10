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
from ftl.php import builder
from ftl.php import layer_builder

_COMPOSER_JSON = json.loads("""
{
  "name": "hello-world",
  "require": {
    "php": ">=5.5",
    "silex/silex": "^1.3"
  },
  "require-dev": {
    "behat/mink": "^1.7",
    "behat/mink-goutte-driver": "^1.2",
    "phpunit/phpunit": "~4",
    "symfony/browser-kit": "^3.0",
    "symfony/http-kernel": "^3.0",
    "google/cloud-tools": "^0.6"
  }
}
""")

_COMPOSER_JSON_TEXT = json.dumps(_COMPOSER_JSON)

_APP = """
require_once __DIR__ . '/../vendor/autoload.php';
$app = new Silex\Application();

$app->get('/', function () {
    return 'Hello World';
});
$app->get('/goodbye', function () {
    return 'Goodbye World';
});

// @codeCoverageIgnoreStart
if (PHP_SAPI != 'cli') {
    $app->run();
}
// @codeCoverageIgnoreEnd

return $app;
"""


class PHPTest(unittest.TestCase):
    @mock.patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def setUp(self, mock_from):
        mock_from.return_value.__enter__.return_value = None
        self._tmpdir = tempfile.mkdtemp()
        self.ctx = context.Memory()
        self.ctx.AddFile("app.php", _APP)
        args = mock.Mock()
        args.name = 'gcr.io/test/test:latest'
        args.base = 'gcr.io/google-appengine/php:latest'
        args.tar_base_image_path = None
        self.builder = builder.PHP(self.ctx, args)
        self.layer_builder = layer_builder.PhaseOneLayerBuilder(
            self.builder._ctx, self.builder._descriptor_files, "/app")

        # Mock out the calls to package managers for speed.
        self.layer_builder._gen_composer_install_tar = mock.Mock()
        self.layer_builder._gen_composer_install_tar.return_value = ('layer',
                                                                     'sha')

    def test_create_package_base_no_descriptor(self):
        self.assertFalse(self.ctx.Contains('composer.json'))
        self.assertFalse(self.ctx.Contains('composer-lock.json'))
        self.layer_builder.BuildLayer()
        lyr = self.layer_builder.GetImage().GetFirstBlob()
        self.assertIsInstance(lyr, str)


if __name__ == '__main__':
    unittest.main()
