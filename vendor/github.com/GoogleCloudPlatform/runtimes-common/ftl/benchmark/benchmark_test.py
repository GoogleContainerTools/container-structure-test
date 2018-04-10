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

from ftl.node.benchmark import main as node_main
from ftl.php.benchmark import main as php_main
from ftl.python.benchmark import main as python_main
import unittest
from mock import patch


class BenchmarkTest(unittest.TestCase):
    @patch('google.cloud.bigquery.Client.create_rows')
    def testPHPBenchmark(self, bigquery_mock):
        php_main.main([
            '--base',
            'gcr.io/gae-runtimes/php72_app_builder:latest',
            '--name',
            'gcr.io/ftl-node-test/benchmark-php-test:test',
            '--directory',
            'ftl/php/benchmark/data/small_app',
            '--iterations',
            '1',
        ])
        assert bigquery_mock.called

    @patch('google.cloud.bigquery.Client.create_rows')
    def testNodeBenchmark(self, bigquery_mock):
        node_main.main([
            '--base',
            'gcr.io/google-appengine/nodejs:latest',
            '--name',
            'gcr.io/ftl-node-test/benchmark-node-test:test',
            '--directory',
            'ftl/node/benchmark/data/small_app',
            '--iterations',
            '1',
        ])
        assert bigquery_mock.called

    @patch('google.cloud.bigquery.Client.create_rows')
    def testPythonBenchmark(self, bigquery_mock):
        python_main.main([
            '--base',
            'gcr.io/google-appengine/python:latest',
            '--name',
            'gcr.io/ftl-node-test/benchmark-python-test:test',
            '--directory',
            'ftl/python/benchmark/data/small_app',
            '--iterations',
            '1',
        ])
        assert bigquery_mock.called


if __name__ == '__main__':
    unittest.main()
