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

"""Reads json files mapping docker digests to tags and reconciles them.
If there are no changes that api call is no-op.
"""

import logging
import os
from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
from containerregistry.transport import transport_pool
import httplib2


class TagReconciler:

    def add_tags(self, digest, tag, dry_run):
        if dry_run:
            logging.debug('Would have tagged {0} with {1}'.format(digest, tag))
            return

        src_name = docker_name.Digest(digest)
        dest_name = docker_name.Tag(tag)
        creds = docker_creds.DefaultKeychain.Resolve(src_name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(src_name, creds, transport) as src_img:
            if src_img.exists():
                creds = docker_creds.DefaultKeychain.Resolve(dest_name)
                logging.debug('Tagging {0} with {1}'.format(digest, tag))
                with docker_session.Push(dest_name, creds, transport) as push:
                        push.upload(src_img)
            else:
                logging.debug("""Unable to tag {0}
                    as the image can't be found""".format(digest))

    def get_existing_tags(self, full_repo, digest):
        full_digest = full_repo + '@sha256:' + digest
        existing_tags = []

        name = docker_name.Digest(full_digest)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(name, creds, transport) as img:
            if img.exists():
                existing_tags = img.tags()
            else:
                logging.debug(
                    """Unable to get existing tags for {0}
                        as the image can't be found""".format(full_digest))
        return existing_tags

    def get_tagged_digest(self, manifests, tag):
        for digest in manifests:
            if tag in manifests[digest]['tag']:
                return digest
        return ''

    def get_digest_from_prefix(self, repo, prefix):
        name = docker_name.Repository(repo)
        creds = docker_creds.DefaultKeychain.Resolve(name)
        transport = transport_pool.Http(httplib2.Http)

        with docker_image.FromRegistry(name, creds, transport) as img:
            digests = [d[len('sha256:'):] for d in img.manifests()]
            matches = [d for d in digests if d.startswith(prefix)]
            if len(matches) == 1:
                return matches[0]
            if len(matches) == 0:
                raise AssertionError('{0} is not a valid prefix'.format(
                                                                 prefix))
        raise AssertionError('{0} is not a unique digest prefix'.format(
                                                                 prefix))

    def reconcile_tags(self, data, dry_run):
        for project in data['projects']:

            default_registry = project['base_registry']
            registries = project.get('additional_registries', [])
            registries.append(default_registry)

            default_repo = os.path.join(default_registry,
                                        project['repository'])

            for image in project['images']:
                digest = self.get_digest_from_prefix(default_repo,
                                                     image['digest'])

                default_digest = default_repo + '@sha256:' + digest
                default_name = docker_name.Digest(default_digest)
                default_creds = (docker_creds.DefaultKeychain
                                 .Resolve(default_name))
                transport = transport_pool.Http(httplib2.Http)

                # Bail out if the digest in the config file doesn't exist.
                with docker_image.FromRegistry(default_name,
                                               default_creds,
                                               transport) as img:

                    if not img.exists():
                        logging.debug('Could not retrieve  ' +
                                      '{0}'.format(default_digest))
                        return

                for registry in registries:

                    full_repo = os.path.join(registry, project['repository'])
                    full_digest = full_repo + '@sha256:' + digest
                    name = docker_name.Digest(full_digest)
                    creds = docker_creds.DefaultKeychain.Resolve(name)

                    with docker_image.FromRegistry(name, creds,
                                                   transport) as img:
                        if img.exists():

                            existing_tags = img.tags()
                            logging.debug('Existing Tags: ' +
                                          '{0}'.format(existing_tags))

                            manifests = img.manifests()
                            tagged_digest = self.get_tagged_digest(
                                manifests, image['tag'])

                            # Don't retag an image if the tag already exists
                            if tagged_digest.startswith('sha256:'):
                                tagged_digest = tagged_digest[len('sha256:'):]
                            if tagged_digest.startswith(digest):
                                logging.debug('Skipping tagging %s with %s as '
                                              'that tag already exists.',
                                              digest, image['tag'])
                                continue

                        # We can safely retag now.
                        full_tag = full_repo + ':' + image['tag']
                        self.add_tags(default_digest, full_tag, dry_run)

                logging.debug(self.get_existing_tags(default_repo, digest))
