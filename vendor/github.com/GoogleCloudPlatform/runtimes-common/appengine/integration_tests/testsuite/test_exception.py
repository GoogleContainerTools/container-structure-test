#!/usr/bin/python

# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import logging
import unittest
import urlparse

import constants
import test_util


class TestException(unittest.TestCase):

    def __init__(self, url, methodName='runTest'):
        self._url = urlparse.urljoin(url, constants.EXCEPTION_ENDPOINT)
        super(TestException, self).__init__()

    def runTest(self):
        payload = test_util.generate_exception_payload()
        _, response_code = test_util.post(self._url, payload)
        self.assertEquals(response_code, 0,
                          'Error encountered inside sample application!')
        logging.info('Token {0} written to Stackdriver '
                     'Error Reporting'.format(payload.get('token')))
