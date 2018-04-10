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
"""This package defines the interface for caching objects."""

import abc
import logging
import datetime

from ftl.common import constants

from containerregistry.client import docker_name
from containerregistry.client import docker_creds
from containerregistry.client.v2_2 import docker_image
from containerregistry.client.v2_2 import docker_session

from ftl.common import ftl_util


class Base(object):
    """Base is an abstract base class representing a layer cache.

    It provides methods used by builders for accessing and storing layers.
    """

    # __enter__ and __exit__ allow use as a context manager.
    @abc.abstractmethod
    def __enter__(self):
        """Initialize the context."""

    def __exit__(self, unused_type, unused_value, unused_traceback):
        """Cleanup the context."""

    @abc.abstractmethod
    def Get(self, cache_key):
        """Lookup a cached image.
        Args:
          cache_key: the cache_key of the package descriptor atop our base.
        Returns:
          the docker_image.Image of the cache hit, or None.
        """

    @abc.abstractmethod
    def Set(self, cache_key, value):
        """Set an entry in the cache.
        Args:
          cache_key: the cache_key of the package descriptor atop our base.
          value: the docker_image.Image to store into the cache.
        """


class Registry(Base):
    """Registry is a cache implementation that stores layers in a registry.

    It stores layers under a 'namespace', with a tag derived from the layer
    cache_key. For example: gcr.io/$repo/$namespace:$cache_key
    """

    def __init__(
            self,
            repo,
            namespace,
            creds,
            transport,
            threads=1,
            should_cache=True,
            should_upload=True,
            mount=None,
            use_global=False,
    ):
        super(Registry, self).__init__()
        self._repo = repo
        self._namespace = namespace
        self._creds = creds
        _reg_name = '{base}/{namespace}'.format(
            base=constants.GLOBAL_CACHE_REGISTRY, namespace=self._namespace)
        # TODO(nkubala): default this to true to point builds to global cache
        self._use_global = use_global
        if use_global:
            _reg = docker_name.Registry(_reg_name)
            self._global_creds = docker_creds.DefaultKeychain.Resolve(_reg)
        self._transport = transport
        self._threads = threads
        self._mount = mount or []
        self._should_cache = should_cache
        self._should_upload = should_upload

    def _tag(self, cache_key, repo=None):
        return docker_name.Tag('{repo}/{namespace}:{tag}'.format(
            repo=repo or str(self._repo),
            namespace=self._namespace,
            tag=cache_key))

    def Get(self, cache_key):
        if not self._should_cache:
            logging.info("--no-cache flag set, cache won't be checked")
            return
        """Attempt to retrieve value from cache."""
        logging.debug('Checking cache for cache_key %s', cache_key)
        hit = self._getEntry(cache_key)
        if hit:
            logging.info('Found cached dependency layer for %s' % cache_key)
            if Registry.checkTTL(hit):
                return hit
            else:
                logging.info(
                    'TTL expired for cached image, rebuilding %s' % cache_key)
        else:
            logging.info('No cached dependency layer for %s' % cache_key)

    def _getEntry(self, cache_key):
        """Retrieve value from cache."""
        # check global cache first
        img = self._getGlobalEntry(cache_key)
        if img:
            return img
        # if we get a global cache miss, check the local cache
        return self._getLocalEntry(cache_key)

    def _getGlobalEntry(self, cache_key):
        if self._use_global:
            key = self._tag(cache_key, constants.GLOBAL_CACHE_REGISTRY)
            entry = Registry.getEntryFromCreds(key, self._global_creds,
                                               self._transport)
            if not entry:
                # TODO(nkubala): standardize this log message so we can
                # crawl cloudbuild logs for cache misses
                logging.info('Cache miss on global cache for %s', key)
            return entry

    def _getLocalEntry(self, cache_key):
        key = self._tag(cache_key)
        entry = Registry.getEntryFromCreds(key, self._creds, self._transport)
        if not entry:
            logging.info('Cache miss on local cache for %s', key)
        return entry

    def Set(self, cache_key, value):
        if not self._should_upload:
            logging.info("--no-upload flag set, images won't be pushed")
            return
        entry = self._tag(cache_key)
        with docker_session.Push(
                entry,
                self._creds,
                self._transport,
                threads=self._threads,
                mount=self._mount) as session:
            session.upload(value)

    @staticmethod
    def getEntryFromCreds(entry, creds, transport):
        """Given a cache entry and a set of credentials authenticated
        to a cache registry, check if the entry exists in the cache."""
        with docker_image.FromRegistry(entry, creds, transport) as img:
            if img.exists():
                logging.info('Found cached base image: %s.' % entry)
                return img
            logging.info('No cached base image found for entry: %s.' % entry)

    @staticmethod
    def checkTTL(entry):
        """Check TTL of cache entry.
        Return whether or not the entry is expired."""
        last_created = ftl_util.timestamp_to_time(
            ftl_util.creation_time(entry))
        now = datetime.datetime.now()
        return last_created > now - datetime.timedelta(
            weeks=constants.DEFAULT_TTL_WEEKS)
