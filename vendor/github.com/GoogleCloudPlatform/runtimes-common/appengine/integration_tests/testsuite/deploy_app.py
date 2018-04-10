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

import datetime
import json
import logging
import os
from retrying import retry
import subprocess
import test_util
import time

from google.cloud import bigquery

import constants
import template

DATASET_NAME = 'cloudperf'
DEPLOY_LATENCY_PROJECT_ENV = 'DEPLOY_LATENCY_PROJECT'
TABLE_NAME = 'deploy_latency'


def _cleanup_files(appdir):
    try:
        os.remove(os.path.join(appdir, 'Dockerfile'))
        os.remove(os.path.join(appdir, 'test.yaml'))
    except Exception:
        pass


def _set_base_image(image):
    # substitute vars in Dockerfile (equivalent of envsubst)
    with open('Dockerfile.in', 'r') as fin:
        with open('Dockerfile', 'w') as fout:
            for line in fin:
                fout.write(line.replace('${STAGING_IMAGE}', image))


def _set_builder_image(builder):
    with open('test.yaml.in', 'r') as fin:
        with open('test.yaml', 'w') as fout:
            for line in fin:
                fout.write(line.replace('${STAGING_BUILDER_IMAGE}', builder))


def _record_latency_to_bigquery(deploy_latency, language, is_xrt):
    current_date = datetime.datetime.now()
    row = [(language, current_date, deploy_latency, is_xrt)]

    project = os.environ.get(DEPLOY_LATENCY_PROJECT_ENV)
    if not project:
        logging.warn('No project specified to record deployment latency!')
        logging.warn('If you wish to record deployment latency, \
                     please set %s env var and try again.',
                     DEPLOY_LATENCY_PROJECT_ENV)
        return 0
    logging.debug('Fetching bigquery client for project %s', project)
    client = bigquery.Client(project=project)
    dataset = client.dataset(DATASET_NAME)
    logging.debug('Writing bigquery data to table %s in dataset %s',
                  TABLE_NAME, dataset)
    table_ref = bigquery.TableReference(dataset_ref=dataset,
                                        table_id=TABLE_NAME)
    table = client.get_table(table_ref)
    return client.create_rows(table, row)


def deploy_app_and_record_latency(appdir, language, is_xrt):
    start_time = time.time()

    version, url = deploy_app(appdir=appdir, environment=constants.GAE)

    # Latency is in seconds round up to 2 decimals
    deploy_latency = round(time.time() - start_time, 2)

    try:
        # Store the deploy latency data to bigquery
        _record_latency_to_bigquery(deploy_latency, language, is_xrt)
    except Exception as e:
        # log error, but ignore and try and cleanup version anyway
        logging.error("Error writing latency to bigquery: %s", e)
    return version, url


def deploy_app_gae(yaml):
    logging.debug('Starting deploy to GAE')

    deployed_version = test_util.generate_version()

    # TODO: once sdk driver is published, use it here
    deploy_command = ['gcloud', 'app', 'deploy', '--no-promote',
                      '--version', deployed_version, '-q']
    if yaml:
        logging.info(yaml)
        deploy_command.append(yaml)

    try:
        test_util.execute_command(deploy_command, True)
        return (deployed_version,
                test_util.retrieve_url_for_version(deployed_version))
    except subprocess.CalledProcessError as e:
        try:
            stop_version(deployed_version)
        except Exception:
            pass
        raise e


def deploy_app_gke(yaml):
    logging.debug('Starting deploy to GKE')

    image_name = test_util.generate_gke_image_name()
    service_name = test_util.generate_gke_service_name()
    namespace = test_util.generate_namespace()

    build_command = ['docker', 'build', '-t', image_name, '.']
    test_util.execute_command(build_command, True)

    push_command = ['gcloud', 'docker', '--', 'push', image_name]
    test_util.execute_command(push_command, True)

    # This command updates the kubeconfig file with credentials for the given
    # GKE cluster. Essentially, points kubectl to our preconfigured cluster.
    cred_command = ['gcloud', 'container', 'clusters',
                    'get-credentials', constants.CLUSTER_NAME]
    test_util.execute_command(cred_command, True)

    namespace_command = ['kubectl', 'create', 'namespace', namespace]
    test_util.execute_command(namespace_command, True)

    deploy_command = ['kubectl', 'apply', '-f', '-',
                      '--namespace', namespace]

    try:
        test_util.execute_command(deploy_command, True,
                                  template.GKE_TEMPLATE.format(
                                    service_name=service_name,
                                    test_image=image_name))

        return namespace, test_util.get_external_ip_for_cluster(service_name,
                                                                namespace)
    except subprocess.CalledProcessError as e:
        try:
            stop_deployment(namespace)
        except Exception:
            pass
        raise e


def deploy_app(appdir, environment, base_image=None,
               builder_image=None, yaml=None):
    try:
        if yaml:
            # convert yaml to absolute path before changing directory
            yaml = os.path.abspath(yaml)

        # change to app directory (and remember original directory)
        owd = os.getcwd()
        os.chdir(appdir)

        # fills in image field in templated Dockerfile and/or builder yaml
        if base_image:
            _set_base_image(base_image)
        if builder_image:
            _set_builder_image(builder_image)

        if environment == constants.GAE:
            return deploy_app_gae(yaml)
        elif environment == constants.GKE:
            return deploy_app_gke(yaml)
        else:
            raise Exception('Invalid environment provided: %s', environment)

    except subprocess.CalledProcessError as cpe:
        logging.error('Error encountered when deploying application! %s',
                      cpe.output)
        raise
    except Exception as e:
        logging.error('Error encountered when deploying application! %s', e)
        raise
    finally:
        _cleanup_files(appdir)
        os.chdir(owd)


@retry(wait_fixed=4000, stop_max_attempt_number=8)
def stop_version(version):
    logging.debug('Removing application version %s', version)
    try:
        delete_command = ['gcloud', 'app', 'services', 'delete',
                          'default', '--version', version, '-q']

        subprocess.check_output(delete_command)
    except subprocess.CalledProcessError as cpe:
        logging.error('Error encountered when deleting app version! %s',
                      cpe.output)
        raise


@retry(wait_fixed=4000, stop_max_attempt_number=8)
def stop_deployment(namespace):
    logging.debug('Removing namespace %s', namespace)
    try:
        service_command = ['kubectl', 'get', 'services', '--namespace',
                           namespace, '--output=json']
        output = test_util.execute_command(service_command, True)
        logging.info(output)
        services = json.loads(output)['items']
        for service in services:
            name = service['metadata']['name']
            delete_service_cmd = ['kubectl', 'delete', 'service', name,
                                  '--namespace', namespace]
            test_util.execute_command(delete_service_cmd, True)
        delete_command = ['kubectl', 'delete', 'namespace', namespace]
        test_util.execute_command(delete_command, True)
    except subprocess.CalledProcessError as cpe:
        logging.error('Error encountered when deleting namespace! '
                      'Manual cleanup may be necessary. %s', cpe.output)
    except Exception as e:
        logging.error('Error encountered when deleting services! '
                      'Manual cleanup may be necessary. %s', e)
