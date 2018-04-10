#!/usr/bin/python

# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import logging
import os
import yaml
import subprocess
import tempfile


RUNTIME_BUCKET = 'runtime-builders'
RUNTIME_BUCKET_PREFIX = 'gs://{0}/'.format(RUNTIME_BUCKET)
MANIFEST_FILE = RUNTIME_BUCKET_PREFIX + 'runtimes.yaml'

SCHEMA_VERSION = 1


def copy_to_gcs(file_path, gcs_path):
    command = ['gsutil', 'cp', file_path, gcs_path]
    try:
        output = subprocess.check_output(command)
        logging.debug(output)
    except subprocess.CalledProcessError as cpe:
        logging.error('Error encountered when writing to GCS! %s', cpe)
    except Exception as e:
        logging.error('Fatal error encountered when shelling command {0}'
                      .format(command))
        logging.error(e)


def write_to_gcs(gcs_path, file_contents):
    try:
        logging.info(gcs_path)
        fd, f_name = tempfile.mkstemp(text=True)
        os.write(fd, file_contents)

        copy_to_gcs(f_name, gcs_path)
    finally:
        os.remove(f_name)


def get_file_from_gcs(gcs_file, temp_file):
    command = ['gsutil', 'cp', gcs_file, temp_file]
    try:
        subprocess.check_output(command, stderr=subprocess.STDOUT)
        return True
    except subprocess.CalledProcessError as e:
        logging.error('Error when retrieving file from GCS! {0}'
                      .format(e.output))
        return False


def load_manifest_file():
    try:
        _, tmp = tempfile.mkstemp(text=True)
        command = ['gsutil', 'cp', MANIFEST_FILE, tmp]
        subprocess.check_output(command, stderr=subprocess.STDOUT)
        with open(tmp) as f:
            return yaml.load(f)
    except subprocess.CalledProcessError:
        logging.info('Manifest file not found in GCS: creating new one.')
        return {'schema_version': SCHEMA_VERSION}
    finally:
        os.remove(tmp)


# 'gsutil ls' would eliminate the try/catch here, but it's eventually
# consistent, while 'gsutil stat' is strongly consistent.
def file_exists(remote_path):
    try:
        logging.info('Checking file {0}'.format(remote_path))
        command = ['gsutil', 'stat', remote_path]
        subprocess.check_call(command, stdout=subprocess.PIPE,
                              stderr=subprocess.PIPE)
        return True
    except subprocess.CalledProcessError:
        return False


class Node:
    def __init__(self, name, isBuilder, child):
        self.name = name
        self.isBuilder = isBuilder
        self.child = child

    def __repr__(self):
        return '{0}: {1}|{2}'.format(self.name, self.isBuilder, self.child)
