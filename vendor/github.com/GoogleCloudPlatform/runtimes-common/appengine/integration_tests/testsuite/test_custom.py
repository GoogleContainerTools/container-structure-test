#!/usr/bin/python

# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the 'License');
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an 'AS IS' BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import re
import json
import logging
import requests
import unittest
import urlparse

import constants
import test_util


class TestCustom(unittest.TestCase):
    """This TestCase fetch a configuration from the endpoint '/custom'
       describing a series of tests, then run each of them and report their
       results.

    In the case where a test of the series fail, this TestCase will be
    considered as failed.

    The specification for the custom tests can be found at:
    https://github.com/GoogleCloudPlatform/runtimes-common/tree/master/integration_tests#custom-tests # noqa
    """

    def __init__(self, url, methodName='runTest'):
        self._base_url = url
        self._url = urlparse.urljoin(url, constants.CUSTOM_ENDPOINT)
        unittest.TestCase.__init__(self)

    def runTest(self):
        """Retrieve the configuration for the custom tests and launch the
           tests.

        Returns:
            None.
        """
        logging.debug('Retrieving list of custom test endpoints.')
        output, status_code = test_util.get(self._url)
        self.assertEquals(status_code, 0,
                          'Cannot connect to sample application!')

        test_num = 0
        if not output:
            logging.debug('No custom tests specified.')
        else:
            for specification in json.loads(output):
                test_num += 1
                self._run_test_for_specification(specification, test_num)

    def _run_test_for_specification(self, specification, test_num):
        """Given the specification for a test execute the steps and
            validate the result.

        Args:
            specification: Dictionary containing the specification.
            test_num: Identifier of the test.

        Returns:
            None.
        """
        name = specification.get('name', 'test_{0}'.format(test_num))
        logging.info('Running custom test: %s', name)

        if self._is_simple_test(specification):
            self._run_simple_test(specification)
        else:
            self._run_composite_test(name, specification)

    def _run_composite_test(self, name, specification):
        """Execute a test composed of multiple step.

        Args:
            name: Name of the test
            specification: Dictionary containing the specification of the test.

        Returns:
            Fail the TestCase if the configuration is invalid or if the
            validation is negative.
        """
        timeout = specification.get('timeout', 2000)
        steps_specification = specification.get('steps', [])
        validation_specification = specification.get('validation')

        if validation_specification is None:
            self.fail('A validation must be specified in the step '
                      'configuration')
        context = {'name': name}
        step_num = 0
        for step in steps_specification:
            step_name, step_context = self._run_step(context, step,
                                                     step_num, timeout)
            context[step_name] = step_context
        logging.debug('context : %s', json.dumps(context,
                                                 sort_keys=True,
                                                 indent=4,
                                                 separators=(',', ': ')))
        self._validate(context, validation_specification)

    def _run_step(self, context, step, step_num, timeout):
        """Use the provided step's configuration to send a request to the
           specified path and store the result into the context.

        Args:
            context: A dictionary containing the context for the test.
            step: A dictionary containing the configuration of the step,
               this include:
                 name (optional): name of the step.
                 configuration (optional):
                    method: 'GET' or 'POST'.
                    headers: Dictionary containing the headers of the request.
                    content: Payload attached to the request.
                 path: Url of the request
            step_num: Index of the step.

        Returns:
            None.
        """
        step_name = step.get('name', 'step_{0}'.format(step_num))
        configuration = step.get('configuration', dict())
        path = step.get('path')

        logging.info('Running step {0} of test {1}, sending request to'
                     'path {2}'.format(step_name, context.get('name'), path))

        test_endpoint = urlparse.urljoin(self._base_url, path)
        response = requests.request(method=configuration.get('method', 'GET'),
                                    url=test_endpoint,
                                    headers=configuration.get('headers'),
                                    data=configuration.get('content'),
                                    timeout=timeout)

        if 'application/json' in response.headers.get('Content-Type'):
            content = response.json()
        else:
            content = response.text

        step_context = {
            'request': {
                'configuration': configuration,
                'path': path
            },
            'response': {
                'headers': dict(response.headers),
                'status': response.status_code,
                'content': content
            }
        }

        return step_name, step_context

    def _validate(self, context, validation_specification):
        """Compare the specification with the context and assert that every key
           present in the specification is also present in the context, and
           that the value associated to that key in the context match the
           regular expression specified in the specification.

        Args:
            context: Dictionary containing for each step the request and
               the response.
            validation_specification: Dictionary with the following fields:
               match: List of object containing:
                 key: Path in the context e.g step.response.headers.property .
                 pattern: Regular expression to be compared with the value
                          present at the path `key` in the context.
        Returns:
            None.
        """
        match = validation_specification.get('match', [])
        endpoint = validation_specification.get('endpoint')
        for test in match:
            key = test.get('key')
            value = self._get_value_at_path(context, key)
            pattern = test.get('pattern')
            self.assertIsNotNone(re.search(pattern, value),
                                 'The value `{0}` for the key `{1}` '
                                 'do not match the pattern `{2}`'
                                 .format(value, key, pattern))

        if endpoint:
            _, response_code = test_util.post(url=urlparse.urljoin(
                                              self._base_url,
                                              endpoint),
                                              payload=context)
            self.assertEqual(response_code, 0, "The validation endpoint {0}"
                                               "returned a negative response"
                                               .format(endpoint))

    def _get_value_at_path(self, context, path):
        """Search for the path `path` in the context and return the associated
           value.

        If the path is not valid the test is considered failed.

        Args:
            context: A dictionary in which the key will be searched.
            path: A list of keys separated by dots, representing a path
                     in the context.
        Returns:
            The value present in the context at the path `path`.
        """
        for key in path.split('.'):
            context = context.get(key)
            self.assertIsNotNone(context, 'An error occurred during the '
                                          'substitution: the key {0} of path '
                                          '{1} is not present in the context'
                                          .format(key, path))
        return context

    def _run_simple_test(self, specification):
        """Send a request to the url specified in the path field of the
           specification.

        Args:
            specification: Dictionary containing the specification for the
                           test.

        Returns:
            In the case where the test is executed but the result is negative
            the TestCase is considered as fail.
        """
        path = specification.get('path')
        timeout = specification.get('timeout', 2000)

        test_endpoint = urlparse.urljoin(self._base_url, path)
        response, status = test_util.get(test_endpoint, timeout=timeout)
        logging.debug(response)
        self.assertEqual(status, 0, 'The response of the endpoint {0} '
                         'is not valid (2xx expected)'.format(path))

    def _is_simple_test(self, specification):
        """Verify if the test specify only a path or a list of steps and a
           validation.

        Args:
            specification: Dictionary containing the specification for the
                           test.

        Returns:
            Whether this is a simple test (containing only a path) or a
            composite test (with multiple steps).
        """
        path = specification.get('path')
        if path is not None:
            if 'steps' in specification or 'validation' in specification:
                self.fail('When the field path is specified, the fields '
                          'validation and steps should not be present')
            return True
