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

from containerregistry.client.v2_2 import docker_image


class MockRegistry():

    def __init__(self):
        self._registry = {}
        self._manifests = {}
        self._tags = {}

    def getRegistry(self):
        return self

    def getFullRepoStr(self, full_repo):
        if isinstance(full_repo, docker_image.FromRegistry):
            full_repo = full_repo.getName()
        return str(full_repo)

    def getRepoStr(self, repository):
        repository = str(repository)
        if ':' in repository:
            repository = repository[:repository.find(':')]
        if '@' in repository:
            repository = repository[:repository.find('@')]
        return repository

    def setTags(self, full_repo, tags):
        if not self.existsImage(full_repo):
            raise AssertionError('{0} does not exist in registry'.format(
                  full_repo))
        self._tags[full_repo] = tags

    def getTags(self, full_repo):
        repo = self.getFullRepoStr(full_repo)
        if repo in self._tags:
            return self._tags[repo]
        raise AssertionError('No tags exist for {0}'.format(repo))

    def setImage(self, full_repo, image):
        full_repo = self.getFullRepoStr(full_repo)
        self._registry[full_repo] = image

    def getImage(self, full_repo):
        full_repo = self.getFullRepoStr(full_repo)
        if full_repo in self._registry:
            return self._registry[full_repo]
        raise AssertionError('{0} does not exist in registry'.format(
                             full_repo))

    def existsImage(self, full_repo):
        full_repo = self.getFullRepoStr(full_repo)
        return full_repo in self._registry

    def setManifests(self, repository, manifest):
        repository = self.getRepoStr(repository)
        self._manifests[repository] = manifest

    def getManifests(self, repository):
        repository = self.getRepoStr(repository)
        if repository in self._manifests:
            return self._manifests[repository]
        raise AssertionError('{0} has no manifest'.format(
                             repository))

    def clearRegistry(self):
        self._registry = {}
