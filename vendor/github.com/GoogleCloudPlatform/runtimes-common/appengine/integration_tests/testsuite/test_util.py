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

import binascii
import json
import logging
import os
import random
import requests
from retrying import retry
import string
from subprocess import Popen, PIPE, check_output, CalledProcessError
import google.auth

import constants

requests.packages.urllib3.disable_warnings()


def _generate_name():
    name = ''.join(random.choice(string.ascii_uppercase +
                   string.ascii_lowercase)
                   for i in range(constants.LOGNAME_LENGTH))
    return name


def _generate_hex_token(num_bytes):
    return binascii.b2a_hex(os.urandom(num_bytes))


def _generate_int64_token():
    return random.randint(-(2 ** 31), (2 ** 31)-1)


def generate_logging_payloads():
    payloads = []
    for s in constants.SEVERITIES:
        payloads.append({
            'log_name': _generate_name(),
            'token': constants.LOGGING_PREFIX + _generate_hex_token(16),
            'level': s
            })
    return payloads


def generate_metrics_payload():
    data = {'name': constants.METRIC_PREFIX.format(_generate_name()),
            'token': _generate_int64_token()}
    return data


def generate_exception_payload():
    data = {'token': _generate_int64_token()}
    return data


def get(url, timeout=constants.DEFAULT_TIMEOUT):
    logging.info('Making GET request to url {0}'.format(url))
    try:
        response = requests.get(url)
        logging.debug('Response: {0}'.format(response.content))
        return _check_response(response,
                               'error when making get ' +
                               'request! url: {0}'
                               .format(url))
    except Exception as e:
        logging.error('Error encountered when making get request!')
        logging.error(e)
        return None, 1


def post(url, payload, timeout=constants.DEFAULT_TIMEOUT):
    try:
        headers = {'Content-Type': 'application/json'}
        response = requests.post(url,
                                 json.dumps(payload),
                                 timeout=timeout,
                                 headers=headers)
        return _check_response(response, 'error when posting request! url: {0}'
                               .format(url))
    except requests.exceptions.Timeout:
        logging.error('POST to {0} timed out after {1} seconds!'
                      .format(url, timeout))
        return 'ERROR', 1


def _check_response(response, error_message):
    if response.status_code - 200 >= 100:  # 2xx
        logging.error('{0} exit code: {1}, text: {2}'
                      .format(error_message,
                              response.status_code,
                              response.text))
        return response.text, 1
    return response.text, 0


def project_id():
    try:
        _, project = google.auth.default()
        return project
    except Exception as e:
        logging.error('Error encountered when retrieving project id!')
        logging.error(e)


def generate_version():
    return 'integration-{0}'.format(_generate_hex_token(8))


@retry(wait_fixed=10000, stop_max_attempt_number=4)
def retrieve_url_for_version(version):
    try:
        # retrieve url of deployed app for test driver
        url_command = ['gcloud', 'app', 'versions', 'describe',
                       version, '--service',
                       'default', '--format=json']
        app_dict = json.loads(check_output(url_command))
        return app_dict.get('versionUrl')
    except (CalledProcessError, ValueError, KeyError) as e:
        logging.error('Error encountered when retrieving app URL! %s', e)
        raise
    raise Exception('Unable to contact deployed application!')


def generate_gke_image_name():
    return 'gcr.io/{project}/{image}'.format(
        project=project_id(),
        image=_generate_hex_token(8)
    )


def generate_gke_service_name():
    return 'gcp-integration-test-{0}'.format(_generate_hex_token(8))


def generate_namespace():
    return 'int-test-{0}'.format(_generate_hex_token(8))


@retry(wait_exponential_multiplier=1000, wait_exponential_max=32000,
       stop_max_attempt_number=12)
def get_external_ip_for_cluster(service_name, namespace):
    logging.info('Waiting for deployment external IP...')
    ip_command = ['kubectl', 'get', 'services', service_name,
                  '--namespace', namespace, '--output=json']
    service = json.loads(execute_command(ip_command))
    ip = service['status']['loadBalancer']['ingress'][0]['ip']
    return 'http://{0}:80'.format(ip)


def get_environment(base_url):
    env_url = base_url + constants.ENVIRONMENT_ENDPOINT
    env, resp_code = get(env_url)
    if not env or resp_code:
        logging.error('Error when retrieving environment from application')
        logging.error('Defaulting to GAE')
        return constants.GAE
    return env


def execute_command(command, print_output=False, stdin=None):
    logging.debug(command)
    proc = Popen(command, shell=False, stdout=PIPE, stderr=PIPE, stdin=PIPE)

    if stdin:
        output, err = proc.communicate(stdin)
    else:
        output, err = proc.communicate()
    exitCode = proc.returncode

    if exitCode != 0:
        raise Exception(err)

    logging.debug(output)
    if print_output:
        logging.info(output)

    return output
