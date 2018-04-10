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
"""This package provides DockerImage for examining docker_build outputs."""

import json

from containerregistry.client.v2_2 import docker_digest
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_http
from containerregistry.transform.v2_2 import metadata as v2_2_metadata


class FromFSImage(docker_image.DockerImage):
    """Interface for implementations that interact with Docker images."""

    def __init__(self, blob_lst, u_layer_lst, overrides={}):
        digest_to_blob, digest_to_u_blob = self._gen_digest_to_blob_and_u_blob(
            blob_lst, u_layer_lst)
        self._digest_to_blob = digest_to_blob
        self._digest_to_u_blob = digest_to_u_blob
        self._diff_id_to_u_layer = self._gen_diff_id_to_u_layer(u_layer_lst)
        self._overrides = overrides
        self._manifest = None
        self._config_file = None

    def GetFirstBlob(self):
        for digest in self._digest_to_blob:
            return self._digest_to_blob[digest]

    def _gen_digest_to_blob_and_u_blob(self, blob_lst, u_layer_lst):
        digest_to_blob = {}
        digest_to_u_blob = {}
        for blob, u_layer in zip(blob_lst, u_layer_lst):
            digest = docker_digest.SHA256(blob)
            digest_to_blob[digest] = blob
            digest_to_u_blob[digest] = u_layer
        return digest_to_blob, digest_to_u_blob

    def _gen_diff_id_to_u_layer(self, u_layer_lst):
        diff_id_to_u_layer = {}
        for u_layer in u_layer_lst:
            diff_id_to_u_layer[docker_digest.SHA256(u_layer)] = u_layer
        return diff_id_to_u_layer

    def fs_layers(self):
        """The ordered collection of filesystem layers that
        comprise this image."""
        manifest = json.loads(self.manifest())
        return [x['digest'] for x in reversed(manifest['layers'])]

    def diff_ids(self):
        """The ordered list of uncompressed layer hashes
        (matches fs_layers)."""
        cfg = json.loads(self.config_file())
        return list(reversed(cfg.get('rootfs', {}).get('diff_ids', [])))

    def config_blob(self):
        manifest = json.loads(self.manifest())
        return manifest['config']['digest']

    def blob_set(self):
        """The unique set of blobs that compose to create the filesystem."""
        return set(self.fs_layers() + [self.config_blob()])

    def digest(self):
        """The digest of the manifest."""
        return docker_digest.SHA256(self.manifest())

    def media_type(self):
        """The media type of the manifest."""
        manifest = json.loads(self.manifest())

        return manifest.get('mediaType', docker_http.OCI_MANIFEST_MIME)

    def manifest(self):
        """The JSON manifest referenced by the tag/digest.

        Returns:
          The raw json manifest
        """
        if self._manifest is None:
            content = self.config_file().encode('utf-8')
            self._manifest = json.dumps(
                {
                    'schemaVersion':
                    2,
                    'mediaType':
                    docker_http.MANIFEST_SCHEMA2_MIME,
                    'config': {
                        'mediaType': docker_http.CONFIG_JSON_MIME,
                        'size': len(content),
                        'digest': docker_digest.SHA256(content)
                    },
                    'layers': [{
                        'mediaType': docker_http.LAYER_MIME,
                        'size': self.blob_size(digest),
                        'digest': digest
                    } for digest in self._digest_to_blob]
                },
                sort_keys=True)
        return self._manifest

    def config_file(self):
        """The raw blob string of the config file."""
        if self._config_file is None:
            _PROCESSOR_ARCHITECTURE = 'amd64'
            _OPERATING_SYSTEM = 'linux'

            entrypoint = self._overrides.pop('Entrypoint', [])
            env = self._overrides.pop('Env', {})
            exposed_ports = self._overrides.pop('ExposedPorts', {})

            output = v2_2_metadata.Override(
                json.loads('{}'),
                v2_2_metadata.Overrides(
                    author='Bazel',
                    created_by='bazel build ...',
                    layers=[k for k in self._diff_id_to_u_layer],
                    entrypoint=entrypoint,
                    env=env,
                    ports=exposed_ports),
                architecture=_PROCESSOR_ARCHITECTURE,
                operating_system=_OPERATING_SYSTEM)
            output['rootfs'] = {
                'diff_ids': [k for k in self._diff_id_to_u_layer]
            }
            if len(self._overrides) > 0:
                output.update(self._overrides)
            self._config_file = json.dumps(output, sort_keys=True)
        return self._config_file

    def blob_size(self, digest):
        """The byte size of the raw blob."""
        return len(self.blob(digest))

    def blob(self, digest):
        """The raw blob of the layer.

        Args:
          digest: the 'algo:digest' of the layer being addressed.

        Returns:
          The raw blob string of the layer.
        """
        return self._digest_to_blob[digest]

    def uncompressed_blob(self, digest):
        """Same as blob() but uncompressed."""
        return self._digest_to_u_blob[digest]

    def _diff_id_to_digest(self, diff_id):
        for (this_digest, this_diff_id) in zip(self.fs_layers(),
                                               self.diff_ids()):
            if this_diff_id == diff_id:
                return this_digest
        raise ValueError('Unmatched "diff_id": "%s"' % diff_id)

    def layer(self, diff_id):
        """Like `blob()`, but accepts the `diff_id` instead.

        The `diff_id` is the name for the digest of the uncompressed layer.

        Args:
          diff_id: the 'algo:digest' of the layer being addressed.

        Returns:
          The raw compressed blob string of the layer.
        """
        return self.blob(self._diff_id_to_digest(diff_id))

    def uncompressed_layer(self, diff_id):
        """Same as layer() but uncompressed."""
        return self._diff_id_to_u_layer[diff_id]
        # return self.uncompressed_blob(self._diff_id_to_digest(diff_id))

    def __enter__(self):
        """Open the image for reading."""

    def __exit__(self, unused_type, unused_value, unused_traceback):
        """Close the image."""

    def __str__(self):
        """A human-readable representation of the image."""
        return str(type(self))
