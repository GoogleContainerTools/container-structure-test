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

"""
Reads json files mapping docker digests to tags and reconciles them.
"""

import argparse
import json
import logging
import os
import shutil
import unittest
import tempfile
from containerregistry.tools import patched
from appengine.reconciletags import tag_reconciler
from appengine.reconciletags import data_integrity_test
from appengine.reconciletags import config_integrity_test


def run_test(files, test):
    # Save original directory
    original_dir = os.getcwd()
    # Create temp directory
    tmpdir = tempfile.mkdtemp()

    # Then, create base and config/tag directories within tmpdir
    base_dir = os.path.join(tmpdir, 'base')
    os.mkdir(base_dir)
    filesdir = os.path.join(tmpdir, 'config/tag/')
    os.makedirs(filesdir)

    # Copy JSON files into config/tag
    for file in files:
        if os.path.isfile(file):
            shutil.copy(file, filesdir)
        else:
            raise AssertionError("{0} is not a valid file".format(file))

    # Switch to base directory and run tests
    os.chdir(base_dir)
    suite = unittest.TestLoader().loadTestsFromTestCase(test)
    unittest.TextTestRunner().run(suite)

    # Return to original directory
    os.chdir(original_dir)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--dry-run', dest='dry_run',
                        help='Runs tests to make sure input files are valid, \
                        and runs a dry run of the reconciler',
                        action='store_true', default=False)
    parser.add_argument('files',
                        help='The files to run the reconciler on',
                        nargs='+')
    parser.add_argument('--data-integrity', dest='data_integrity',
                        help='Runs a test to make sure the data in the input \
                        files is the same as in prod',
                        action='store_true', default=False)
    args = parser.parse_args()
    logging.basicConfig(level=logging.DEBUG)

    if args.data_integrity:
        run_test(args.files, data_integrity_test.DataIntegrityTest)
        return

    if args.dry_run:
        run_test(args.files, config_integrity_test.ReconcilePresubmitTest)

    r = tag_reconciler.TagReconciler()
    for f in args.files:
        logging.debug('---Processing {0}---'.format(f))
        with open(f) as tag_map:
            data = json.load(tag_map)
            r.reconcile_tags(data, args.dry_run)


if __name__ == '__main__':
    with patched.Httplib2():
        main()
