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

DEFAULT_LOG_LEVEL = 'NOTSET'

DEFAULT_DESTINATION_PATH = 'srv'
DEFAULT_ENTRYPOINT = None

# docker transport thread config
THREADS = 32

# cache constants
DEFAULT_TTL_WEEKS = 1
GLOBAL_CACHE_REGISTRY = 'gcr.io/ftl-global-cache'

# node constants
NODE_NAMESPACE = 'node-package-lock-cache'
PACKAGE_LOCK = 'package-lock.json'
PACKAGE_JSON = 'package.json'
NODE_DEFAULT_ENTRYPOINT = 'node server.js'
NPMRC = '.npmrc'

# python constants
PIPFILE_LOCK = 'Pipfile.lock'
PIPFILE = 'Pipfile'
REQUIREMENTS_TXT = 'requirements.txt'
PYTHON_NAMESPACE = 'python-requirements-cache'
VENV_DIR = '/env'
WHEEL_DIR = 'wheel'

# logging constants
PHASE_1_CACHE_STR = '{key_version}:{language}->{key}'
PHASE_2_CACHE_STR = '{key_version}:{language}:{package_name}:' \
            '{package_version}->{key}'
CACHE_HIT = '[CACHE][HIT] '
CACHE_MISS = '[CACHE][MISS] '

PHASE_1_CACHE_HIT = CACHE_HIT + PHASE_1_CACHE_STR
PHASE_2_CACHE_HIT = CACHE_HIT + PHASE_2_CACHE_STR
PHASE_1_CACHE_MISS = CACHE_MISS + PHASE_1_CACHE_STR
PHASE_2_CACHE_MISS = CACHE_MISS + PHASE_2_CACHE_STR

CACHE_KEY_VERSION = 'v1'
