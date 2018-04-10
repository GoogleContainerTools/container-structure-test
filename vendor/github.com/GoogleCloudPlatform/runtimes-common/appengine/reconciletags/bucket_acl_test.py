"""Bucket ACL test.

Looks at the GCS buckets where all GCR container information is stored and
makes sure they're all world readable."""

import json
import logging
import subprocess
import unittest


class BucketAclTest(unittest.TestCase):

    def _get_acls(self, bucket):
        acls = json.loads(
                  subprocess.check_output(['gsutil', 'acl',
                                           'get', bucket]))
        return acls

    def _get_bucket_name(self, bucket_name, mirror):
        if mirror:
            bucket_name = '{0}.{1}'.format(mirror, bucket_name)
        return 'gs://{0}'.format(bucket_name)

    def test_bucket_acls(self):
        repos = ['artifacts.gcp-runtimes.appspot.com',
                 'artifacts.google-appengine.appspot.com',
                 'runtime-builders']
        mirrors = ['', 'asia', 'eu', 'us']
        bad_buckets = []
        for repo in repos:
            for mirror in mirrors:
                if repo == 'runtime-builders' and mirror != '':
                    continue
                # Construct bucket name
                bucket_name = self._get_bucket_name(repo, mirror)

                # Grab acls json
                acls = self._get_acls(bucket_name)

                # Make sure its world readable
                valid = False
                for acl in acls:
                    if acl['entity'] == 'allUsers':
                        if acl['role'] == 'READER':
                            valid = True
                            break
                if not valid:
                    bad_buckets.append(bucket_name)

        if len(bad_buckets) > 0:
            self.fail('The following buckets have incorrect ACLs:'
                      + str(bad_buckets))


if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    unittest.main()
