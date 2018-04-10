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

import sys
import subprocess
import argparse


def check_if_tag_exists(raw_image_path, force_build):
    # extract both path to image, and the tag if provided
    image_parts = raw_image_path.split(':')
    image_path = image_parts[0]
    if len(image_parts) == 2:
        image_tag = image_parts[1]
    else:
        image_tag = 'latest'

    p = subprocess.Popen(["gcloud container images list-tags "
                          + "--format='value(tags)' --no-show-occurrences {0}"
                          .format(image_path)],
                         shell=True, stdout=subprocess.PIPE,
                         stderr=subprocess.STDOUT)

    output, error = p.communicate()
    if p.returncode != 0:
        sys.exit('Error encountered when retrieving existing image tags! '
                 + 'Full log: \n\n' + output)

    existing_tags = set(tag.rstrip() for tag in output.split('\n'))
    print 'Existing tags for image {0}:'.format(image_path)
    for tag in existing_tags:
        print tag

    if image_tag in existing_tags:
        print "Tag '{0}' already exists in remote repository!" \
              .format(image_tag)
        if not force_build:
            sys.exit('Exiting build.')
        else:
            print "Forcing build. Tag '{0}' " \
                  "will be overwritten!".format(image_tag)
            return
    print "Tag '{0}' does not exist in remote repository! " \
          "Continuing with build.".format(image_tag)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--image', type=str,
                        help='Fully qualified remote path for the '
                        + 'target image')
    parser .add_argument('--force', action='store_true', default=False)
    args = parser.parse_args()

    if args.image is None:
        sys.exit('Please provide fully qualified remote path for the '
                 'target image.')

    check_if_tag_exists(args.image, args.force)


if __name__ == "__main__":
    main()
