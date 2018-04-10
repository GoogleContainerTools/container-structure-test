"""Data Integrity tests.

Checks the json config files currently submitted and compares the entries to
what is currently on GCR. Fails if a discrepency is found."""


import glob
import json
import logging
import os
import unittest
from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.client.v2_2 import docker_image
from containerregistry.transport import transport_pool
from containerregistry.tools import patched
import httplib2


class DataIntegrityTest(unittest.TestCase):

    def _get_digests(self, repo):
        name = docker_name.Repository(repo)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(name, creds, transport) as img:
            digests = [d[len('sha256:'):] for d in img.manifests()]
            return digests
        raise AssertionError('Unable to get digests from {0}'.format(repo))

    def _get_tags(self, repo, digest):
        full_digest = repo + '@sha256:' + digest

        name = docker_name.Digest(full_digest)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(name, creds, transport) as img:
            return img.tags()
        raise AssertionError('Unable to get tags from {0}'.format(full_digest))

    def test_data_consistency(self):
        failed_digests = []
        for f in glob.glob('../config/tag/*.json'):
            logging.debug('Testing {0}'.format(f))
            with open(f) as tag_map:
                data = json.load(tag_map)
                for project in data['projects']:
                    full_repo = os.path.join(project['base_registry'],
                                             project['repository'])
                    real_digests = self._get_digests(full_repo)
                    for image in project['images']:
                        for digest in real_digests:
                            if digest.startswith(image['digest']):
                                real_tags = self._get_tags(full_repo, digest)
                                if image['tag'] not in real_tags:
                                    failed_digests.append({full_repo: image})

        if len(failed_digests) > 0:
            self.fail('These entries do not correspond with what is'
                      ' currently live:' + str(failed_digests))


if __name__ == '__main__':
    with patched.Httplib2():
        logging.basicConfig(level=logging.DEBUG)
        unittest.main()
