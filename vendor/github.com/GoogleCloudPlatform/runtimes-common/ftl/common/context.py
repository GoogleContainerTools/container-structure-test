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
"""This package defines the interface for accessing the build context."""

import abc
import os


class Base(object):
    """Base is an abstract base class representing an application context.

    It provides methods used by builders for accessing app files.
    """

    # __enter__ and __exit__ allow use as a context manager.
    def __enter__(self):
        """Initialize the context."""
        return self

    def __exit__(self, unused_type, unused_value, unused_traceback):
        """Cleanup the context."""

    @abc.abstractmethod
    def Contains(self, relative_path):
        """Whether the application context contains the given file."""

    @abc.abstractmethod
    def ListFiles(self):
        """Recursively enumerate the files under the workspace.
        Yields:
          the paths of files within the context, suitable for use with GetFile.
        """

    @abc.abstractmethod
    def GetFile(self, relative_path):
        """Retrieve the contents of a particular file.
        Args:
          relative_path: The relative path of the file to read.
        Returns:
          the contents of the file.
        """


class Workspace(Base):
    """Workspace is a context implementation for working with files in a local
    directory.
    """

    def __init__(self, directory):
        super(Workspace, self).__init__()
        self._directory = directory

    def Contains(self, relative_path):
        """Override."""
        fqpath = os.path.join(self._directory, relative_path)
        return os.path.isfile(fqpath)

    def ListFiles(self):
        """Override."""
        dir = self._directory + '/'
        for root, dirnames, filenames in os.walk(dir):
            relative = root[len(dir):]
            for fname in filenames:
                yield os.path.join(relative, fname)

    def GetFile(self, filename):
        """Override."""
        fqname = os.path.join(self._directory, filename)
        with open(fqname, 'rb') as f:
            return f.read()


class Memory(Base):
    """Memory is a context implementation for working with files in memory."""

    def __init__(self):
        self._files = {}

    def AddFile(self, filename, contents):
        self._files[filename] = contents

    def Contains(self, relative_path):
        return relative_path in self._files

    def ListFiles(self):
        for path in self._files:
            yield path

    def GetFile(self, filename):
        return self._files[filename]
