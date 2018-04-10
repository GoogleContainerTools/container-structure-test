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
from datetime import datetime
import json
import logging
import os
import re
from ruamel import yaml
import subprocess
import sys

import builder_util


IMAGE_REGEX = '((us.|eu.|asia.)?gcr\.io/).*'


def main():
    logging.getLogger().setLevel(logging.INFO)
    parser = argparse.ArgumentParser()
    parser.add_argument('--filename', '-f',
                        help='templated cloudbuild config file.',
                        required=True)
    parser.add_argument('--bucket', '-b',
                        help='GCS bucket to publish runtime to',
                        default='runtime-builders')
    args = parser.parse_args()

    return _resolve_and_publish(args.filename, args.bucket)


def _resolve_and_publish(config_file, bucket):
    try:
        gcs_paths = []
        with open(config_file, 'r') as f:
            if 'yaml' not in config_file:
                logging.error('Please provide a valid yaml config file.')
                sys.exit(1)
            project_cfg = yaml.round_trip_load(f, preserve_quotes=True)
            project_name = project_cfg['project']
            for builder in project_cfg['builders']:
                cfg = os.path.abspath(str(builder['file']))
                name = builder['name']
                builder_name = project_name + '-' + name

                templated_file = _resolve_tags(cfg)
                logging.info(templated_file)
                gcs_paths.append(_publish_to_gcs(templated_file,
                                                 builder_name,
                                                 bucket))

        logging.info('Published Runtimes:')
        logging.info(gcs_paths)
    except ValueError as ve:
        logging.error('Error when parsing config! Check file formatting. \n{0}'
                      .format(ve))
    except KeyError as ke:
        logging.error('Config file is missing required field! \n{0}'
                      .format(ke))


def _resolve_tags(config_file):
    """
    Given a templated YAML cloudbuild config file, parse it, resolve image tags
    on each build step's image to the corresponding digest, and write new
    config with fully qualified images to temporary file for upload to GCS.

    Keyword arguments:
    config_file -- string representing path to
    templated cloudbuild YAML config file

    Return value:
    path to temporary file containing fully qualified config file, to be
    published to GCS.
    """
    with open(config_file, 'r') as infile:
        logging.info('Templating file: {0}'.format(config_file))
        try:
            config = yaml.round_trip_load(infile, preserve_quotes=True)

            for step in config.get('steps'):
                image = step.get('name')
                step['name'] = _resolve_tag(image)
                args = step.get('args', [])
                for i in range(0, len(args)):
                    arg = args[i]
                    m = re.search(IMAGE_REGEX, arg)
                    if m:
                        suffix = m.group()
                        prefix = re.sub(suffix, '', arg)
                        args[i] = prefix + _resolve_tag(suffix)

            return yaml.round_trip_dump(config)
        except yaml.YAMLError as e:
            logging.error(e)
            sys.exit(1)


def _resolve_tag(image):
    """
    Given a path to a tagged Docker image in GCR, replace the tag with its
    corresponding sha256 digest.
    """
    if ':' not in image:
        logging.error('Image \'{0}\' must contain explicit tag or '
                      'digest!'.format(image))
        sys.exit(1)
    elif '@sha256' in image:
        return image
    else:
        parts = image.split(':')
        base_image = parts[0]
        target_tag = parts[1]

    command = ['gcloud', 'container', 'images',
               'list-tags', base_image, '--format=json', '--limit=20']

    try:
        output = subprocess.check_output(command)
        entries = json.loads(output)
        for image in entries:
            for tag in image.get('tags'):
                if tag == target_tag:
                    digest = image.get('digest')
                    return base_image + '@' + digest
        logging.error('Tag {0} not found on image {1}!'
                      .format(target_tag, base_image))
        sys.exit(1)
    except subprocess.CalledProcessError as e:
        logging.error(e)

    logging.error('No digest found for tag {0} on '
                  'image {1}'.format(target_tag, base_image))


def _publish_to_gcs(builder_file_contents, builder_name, bucket):
    """
    Given a cloudbuild YAML config file, publish the file to a bucket in GCS.
    """
    file_name = '{0}-builder-{1}.yaml'.format(
        builder_name,
        datetime.now().strftime('%Y%m%d%H%M%S'))

    full_path = 'gs://{0}/{1}'.format(bucket, file_name)

    builder_util.write_to_gcs(full_path, builder_file_contents)

    return full_path


if __name__ == '__main__':
    sys.exit(main())
