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

from containerregistry.client import docker_creds
from containerregistry.client import docker_name
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session
from containerregistry.transport import transport_pool
import httplib2

DIGEST = '0000000000000000000000000000000000000000000000000000000000000000'


def main():
    digest = 'fake.gcr.io/test/test@sha256:' + DIGEST
    tag = 'fake.gcr.io/test/test:tag'
    src_name = docker_name.Digest(digest)
    dest_name = docker_name.Tag(tag)
    creds = docker_creds.DefaultKeychain.Resolve(src_name)
    transport = transport_pool.Http(httplib2.Http)

    with docker_image.FromRegistry(src_name, creds, transport) as src_img:
        if src_img.exists():
            creds = docker_creds.DefaultKeychain.Resolve(dest_name)
            with docker_session.Push(dest_name, creds, transport) as push:
                    push.upload(src_img)


if __name__ == '__main__':
    main()
