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
"""This package defines abstract methods for
building individual image layers."""

import abc
import hashlib

from ftl.common import constants


class BaseLayerBuilder(object):
    """BaseLayerBuilder is an abstract base class representing
    a builder for a container layer.

    It provides methods for generating a dependency layer and an application
    layer.
    """

    __metaclass__ = abc.ABCMeta  # For enforcing that methods are overriden.

    def __init__(self):
        self._img = None

    def GetImage(self):
        if self._img is None:
            raise Exception('error: layer image was not built yet so ' +
                            'image cannot be accessed')
        return self._img

    def SetImage(self, img):
        self._img = img

    @abc.abstractmethod
    def BuildLayer(self):
        """Synthesizes the application layer.
        Modifies:
          self._img
        """


class CacheableLayerBuilder(BaseLayerBuilder):

    __metaclass__ = abc.ABCMeta  # For enforcing that methods are overriden.

    @abc.abstractmethod
    def GetCacheKeyRaw(self):
        """
        Returns:
          the raw value for the cache key (not hashed)
        """

    def GetCacheKey(self):
        fingerprint = "%s %s" % (self.GetCacheKeyRaw(),
                                 constants.CACHE_KEY_VERSION)
        return hashlib.sha256(fingerprint).hexdigest()

    @abc.abstractmethod
    def BuildLayer(self):
        """Synthesizes the application layer.
        Modifies:
          self._img
        """
