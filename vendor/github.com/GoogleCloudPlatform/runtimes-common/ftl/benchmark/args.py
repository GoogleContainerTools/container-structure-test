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
"""This package defines the shared cli args for ftl binaries."""

from ftl.common import args


def base_parser():
    parser = args.base_parser()

    parser.add_argument(
        '--iterations',
        action='store',
        type=int,
        default=5,
        help='Number of times to build the image')

    parser.add_argument(
        '--description',
        action='store',
        help=('Description of the app being benchmarked.'))

    parser.add_argument(
        '--project',
        action='store',
        default='ftl-node-test',
        help='Bigquery project build times should be stored in')

    parser.add_argument(
        '--dataset',
        action='store',
        default='ftl_benchmark',
        help='Bigquery dataset build times should be stored in')

    parser.add_argument(
        '--gen_files',
        action='store',
        type=int,
        default=0,
        help=('Number of app files to generate for test'))

    return parser
