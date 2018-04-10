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

"""End to end test for the tag reconciler."""

import json
import tag_reconciler
import subprocess
import unittest


class ReconciletagsE2eTest(unittest.TestCase):

    _FILE_NAME = 'e2e_test.json'
    _DIR = 'reconciletags/tiny_docker_image/'
    _REPO = 'gcr.io/gcp-runtimes/reconciler-e2etest'
    _TAG = 'initial'
    _TEST_JSON = {
      "projects": [
        {
          "base_registry": "gcr.io",
          "additional_registries": [],
          "repository": "gcp-runtimes/reconciler-e2etest",
          "images": [
                  {
                      "digest": "",
                      "tag": "testing"
                  }
          ]
        }
      ]
    }

    def _ListTags(self, repo):
        output = json.loads(
            subprocess.check_output(['gcloud', 'container',
                                     'images', 'list-tags',
                                     '--format=json', repo]))
        return output

    def _BuildImage(self, full_image_name):
        # create a non-functional but tiny docker image
        subprocess.call(['gcloud', 'container', 'builds', 'submit', '--config',
                        'appengine/reconciletags/e2e_cloudbuild.yaml', '.'])
        # grab the just created digest
        output = self._ListTags(self._REPO)
        self.assertEqual(len(output), 1)
        output = output.pop()
        self.digest = output['digest'].split(':')[1]
        self.assertEqual(len(output['tags']), 1)
        # write the proper json to the config file
        self._TEST_JSON['projects'][0]['images'][0]['digest'] = self.digest

    def setUp(self):
        self.r = tag_reconciler.TagReconciler()
        self._BuildImage(self._REPO + ':' + self._TAG)

    def tearDown(self):
        subprocess.call(['gcloud', 'container', 'images',
                         'delete', self._REPO + '@sha256:' + self.digest,
                         '-q', '--force-delete-tags'])

    def checkListTagsOutput(self):
        output = self._ListTags(self._REPO)
        for image in output:
            if image['digest'].split(':')[1] == self.digest:
                self.assertEquals(len(image['tags']), 2)
                self.assertEquals(image['tags'][1], 'testing')
                self.assertEquals(image['tags'][0], self._TAG)

    def testTagReconciler(self):
        # run the reconciler
        self.r.reconcile_tags(self._TEST_JSON, False)
        # check list-tags to see if it added the correct tag
        self.checkListTagsOutput()

        # run reconciler again and make sure nothing changed
        self.r.reconcile_tags(self._TEST_JSON, False)
        self.checkListTagsOutput()

        # now try with a fake digest
        self._TEST_JSON['projects'][0]['images'][0]['digest'] = 'fakedigest'
        with self.assertRaises(AssertionError):
            self.r.reconcile_tags(self._TEST_JSON, False)

    def testParFile(self):
        subprocess.call(['./appengine/reconciletags/reconciletags.par',
                        'appengine/reconciletags/'+self._FILE_NAME])
        # check list-tags to see if it added the correct tag
        self.checkListTagsOutput()

        # run reconciler again and make sure nothing changed
        subprocess.call(['./appengine/reconciletags/reconciletags.par',
                        'appengine/reconciletags/'+self._FILE_NAME])
        self.checkListTagsOutput()


if __name__ == '__main__':
    unittest.main()
