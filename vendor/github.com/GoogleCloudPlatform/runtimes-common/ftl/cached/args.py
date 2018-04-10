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
        '--label-1',
        dest='label_1',
        action='store',
        default='original',
        help='image label for original uploaded image')

    parser.add_argument(
        '--label-2',
        dest='label_2',
        action='store',
        default='reupload',
        help='image label for reuploades image')

    return parser
