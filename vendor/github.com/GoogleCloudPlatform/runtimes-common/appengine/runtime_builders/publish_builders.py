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

import argparse
import glob
import logging
import os
import yaml
import sys

import builder_util
import verify_manifest


def main():
    logging.getLogger().setLevel(logging.INFO)

    parser = argparse.ArgumentParser()
    parser.add_argument('--manifest', '-m',
                        help='path to runtime.yaml manifest',
                        required=True)
    parser.add_argument('--directory', '-d',
                        help='path to builder config directory for '
                             'publishing latest')
    args = parser.parse_args()

    try:
        if not args.manifest.endswith('.yaml'):
            logging.error('Please provide path to runtimes.yaml manifest.')
        verify_manifest.verify_manifest(args.manifest)
        builder_util.copy_to_gcs(args.manifest, builder_util.MANIFEST_FILE)

        if args.directory:
            _publish_latest(args.directory)
    except ValueError as ve:
        logging.error('Error when parsing yaml! Check file formatting. \n{0}'
                      .format(ve))
    except KeyError as ke:
        logging.error('Config file is missing required field! \n{0}'
                      .format(ke))


def _publish_latest(builder_dir):
    for f in glob.glob(os.path.join(builder_dir, '*.yaml')):
        with open(f, 'r') as f:
            config = yaml.safe_load(f)

        if 'latest' not in config or 'project' not in config:

            logging.error('Config file {0} is missing a required field!'
                          .format(f))
            continue

        latest = config['latest']
        project_name = config['project']

        parts = os.path.splitext(latest)
        prefix = builder_util.RUNTIME_BUCKET_PREFIX + project_name + '-'
        latest_file = parts[0].replace(prefix, '')
        file_name = '{0}.version'.format(project_name)
        full_path = builder_util.RUNTIME_BUCKET_PREFIX + file_name
        builder_util.write_to_gcs(full_path, latest_file)


if __name__ == '__main__':
    sys.exit(main())
