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

import cStringIO
import datetime
import os
import tarfile
import logging
import subprocess

from ftl.common import constants
from ftl.common import ftl_util
from ftl.common import single_layer_image
from ftl.common import tar_to_dockerimage


class AppLayerBuilder(single_layer_image.BaseLayerBuilder):
    def __init__(self,
                 ctx,
                 destination_path=constants.DEFAULT_DESTINATION_PATH,
                 entrypoint=constants.DEFAULT_ENTRYPOINT,
                 exposed_ports=None):
        self._ctx = ctx
        self._destination_path = destination_path
        self._entrypoint = entrypoint
        self._exposed_ports = exposed_ports

    def GetCacheKeyRaw(self):
        return None

    def BuildLayer(self):
        """Override."""
        with ftl_util.Timing('Building app layer'):
            buf = cStringIO.StringIO()
            logging.info('Starting to generate app layer \
                tarfile from context...')
            with tarfile.open(fileobj=buf, mode='w') as out:
                for name in self._ctx.ListFiles():
                    content = self._ctx.GetFile(name)
                    info = tarfile.TarInfo(
                        os.path.join(self._destination_path.strip("/"), name))
                    info.size = len(content)
                    out.addfile(info, fileobj=cStringIO.StringIO(content))
            logging.info('Finished generating app layer tarfile from context.')

            tar = buf.getvalue()

            logging.info('Starting to gzip app layer tarfile...')
            gzip_process = subprocess.Popen(
                ['gzip', '-f'],
                stdout=subprocess.PIPE,
                stdin=subprocess.PIPE,
                stderr=subprocess.PIPE)
            gz = gzip_process.communicate(input=tar)[0]
            overrides_dct = {
                'created': str(datetime.date.today()) + 'T00:00:00Z'
            }
            if self._entrypoint:
                overrides_dct['Entrypoint'] = self._entrypoint
            if self._exposed_ports:
                overrides_dct['ExposedPorts'] = self._exposed_ports
            logging.info('Finished gzipping tarfile.')
            self._img = tar_to_dockerimage.FromFSImage([gz], [tar],
                                                       overrides_dct)
