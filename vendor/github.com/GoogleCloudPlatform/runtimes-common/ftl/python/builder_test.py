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

import unittest
import datetime
import mock
import json

from ftl.common import context
from ftl.common import constants
from ftl.common import ftl_util
from ftl.python import builder
from ftl.python import layer_builder

_REQUIREMENTS_TXT = """
Flask==0.12.0
"""

_APP = """
import os
from flask import Flask
app = Flask(__name__)


@app.route("/")
def hello():
    return "Hello from Python!"


if __name__ == "__main__":
    port = int(os.environ.get("PORT", 5000))
    app.run(host='0.0.0.0', port=port)
"""


class PythonTest(unittest.TestCase):
    @mock.patch('containerregistry.client.v2_2.docker_image.FromRegistry')
    def setUp(self, mock_from):
        mock_from.return_value.__enter__.return_value = None
        self.ctx = context.Memory()
        self.ctx.AddFile("app.py", _APP)
        args = mock.Mock()
        args.name = 'gcr.io/test/test:latest'
        args.base = 'gcr.io/google-appengine/python:latest'
        args.python_cmd = 'python2.7'
        args.pip_cmd = 'pip'
        args.venv_cmd = 'virtualenv'
        args.tar_base_image_path = None
        self.builder = builder.Python(self.ctx, args)

        # constants.VENV_DIR.replace('/', '') is used as the default path
        # will give permissions errors in some build environments (eg: kokoro)
        self.interpreter_builder = layer_builder.InterpreterLayerBuilder(
            ftl_util.gen_tmp_dir(constants.VENV_DIR.replace('/', '')),
            self.builder._python_cmd,
            self.builder._venv_cmd)
        self.interpreter_builder._setup_venv = mock.Mock()
        self.builder._pip_install = mock.Mock()

    def test_build_interpreter_layer_ttl_written(self):
        self.interpreter_builder.BuildLayer()
        overrides = ftl_util.CfgDctToOverrides(
            json.loads(self.interpreter_builder.GetImage().config_file()))

        self.assertNotEqual(overrides.creation_time, "1970-01-01T00:00:00Z")
        last_created = ftl_util.timestamp_to_time(overrides.creation_time)
        now = datetime.datetime.now()
        self.assertTrue(last_created > now - datetime.timedelta(days=2))


if __name__ == '__main__':
    unittest.main()
