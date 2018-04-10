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
from retrying import retry
import unittest
import urlparse
import time

import google.cloud.monitoring

import constants
import test_util


class TestMonitoring(unittest.TestCase):
    def __init__(self, url, methodName='runTest'):
        self._url = urlparse.urljoin(url, constants.MONITORING_ENDPOINT)
        super(TestMonitoring, self).__init__()

    def runTest(self):
        payload = test_util.generate_metrics_payload()
        metric_name, target = payload.get('name'), payload.get('token')
        client = google.cloud.monitoring.Client()

        try:
            _, response_code = test_util.post(self._url, payload,
                                              constants.METRIC_TIMEOUT)
            self.assertEquals(response_code, 0,
                              'Error encountered inside sample application!')

            logging.info('trying to find {0} stored in {1}...'
                         .format(target, metric_name))
            start = time.time()
            try:
                found_token = self._read_metric(metric_name, target, client)
                failure = None
            except Exception as e:
                found_token = False
                failure = e
            elapsed = time.time() - start

            logging.info('time elapsed checking for metric: {0}s'
                         .format(elapsed))

            self.assertTrue(found_token,
                            'Token not found in Stackdriver monitoring!\n'
                            'Last error: {0}'.format(failure))
        finally:
            self._try_cleanup_metric(client, metric_name)

    def _try_cleanup_metric(self, client, metric_name):
        try:
            descriptor = client.metric_descriptor(metric_name)
            descriptor.delete()
            logging.info('metric {0} deleted'.format(metric_name))
        except Exception as e:
            logging.warning('Error when deleting metric {0},'
                            ' manual cleanup might be needed: {1}'
                            .format(metric_name, e.message))

    @retry(wait_fixed=5000, stop_max_attempt_number=20)
    def _read_metric(self, name, target, client):
        query = client.query(name, minutes=5)
        if self._no_timeseries_in(query):
            raise Exception('No timeseries match the query for metric {0}'
                            .format(name))

        for timeseries in query:
            for point in timeseries.points:
                if point.value == target:
                    logging.info('Token {0} found in Stackdriver metrics {1}'
                                 .format(target, name))
                    return True
                print(point.value)
        raise Exception('Token {0} not found in metric {1}'
                        .format(target, name))

    def _no_timeseries_in(self, query):
        if query is None:
            logging.info('query is none')
            return True
        # query is a generator, so sum over it to get the length
        query_length = sum(1 for timeseries in query)
        if query_length == 0:
            logging.info('query is empty')
            return True
        return False
