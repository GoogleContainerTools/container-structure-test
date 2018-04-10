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
"""This package defines helpful utilities for FTL ."""
import os
import time
import logging
import subprocess
import tempfile
import datetime
import json

from containerregistry.client.v2_2 import append
from containerregistry.transform.v2_2 import metadata


def AppendLayersIntoImage(imgs):
    with Timing('Stitching layers into final image'):
        for i, img in enumerate(imgs):
            if i == 0:
                result_image = img
                continue
            diff_ids = img.diff_ids()
            for diff_id in diff_ids:
                lyr = img.blob(img._diff_id_to_digest(diff_id))
                overrides = CfgDctToOverrides(json.loads(img.config_file()))
                result_image = append.Layer(
                    result_image, lyr, diff_id=diff_id, overrides=overrides)
        return result_image


# This is a 'whitelist' of values to pass from the
# config_file of a DockerImage to an Overrides object
# _OVERRIDES_VALUES = ['created', 'Entrypoint', 'Env']
def CfgDctToOverrides(config_dct):
    """
    Takes a dct of config values and runs them through
    the whitelist
    """
    overrides_dct = {}
    for k, v in config_dct.iteritems():
        if k == 'created':
            # this key change is made as the key is
            # 'creation_time' in an Overrides object
            # but 'created' in the config_file
            overrides_dct['creation_time'] = v
    for k, v in config_dct['config'].iteritems():
        if k == 'Entrypoint':
            # this key change is made as the key is
            # 'entrypoint' in an Overrides object
            # but 'Entrypoint' in the config_file
            overrides_dct['entrypoint'] = v
        elif k == 'Env':
            # this key change is made as the key is
            # 'env' in an Overrides object
            # but 'Env' in the config_file
            overrides_dct['env'] = v
        elif k == 'ExposedPorts':
            # this key change is made as the key is
            # 'ports' in an Overrides object
            # but 'ExposedPorts' in the config_file
            overrides_dct['ports'] = v
    return metadata.Overrides(**overrides_dct)


class Timing(object):
    def __init__(self, descriptor):
        self.descriptor = descriptor

    def __enter__(self):
        self.start = time.time()
        return self

    def __exit__(self, unused_type, unused_value, unused_traceback):
        end = time.time()
        logging.info('%s took %d seconds', self.descriptor, end - self.start)


def zip_dir_to_layer_sha(pkg_dir):
    tar_path = tempfile.mktemp(suffix='.tar')
    with Timing('tar_runtime_package'):
        subprocess.check_call(['tar', '-C', pkg_dir, '-cf', tar_path, '.'])

    u_blob = open(tar_path, 'r').read()
    # We use gzip for performance instead of python's zip.
    with Timing('gzip_runtime_tar'):
        subprocess.check_call(['gzip', tar_path, '-1'])
    return open(os.path.join(pkg_dir, tar_path + '.gz'), 'rb').read(), u_blob


def has_pkg_descriptor(descriptor_files, ctx):
    for f in descriptor_files:
        if ctx.Contains(f):
            return True
    return False


def descriptor_parser(descriptor_files, ctx):
    descriptor = None
    for f in descriptor_files:
        if ctx.Contains(f):
            descriptor = f
            descriptor_contents = ctx.GetFile(descriptor)
            break
    if not descriptor:
        logging.info("No package descriptor found. No packages installed.")
        return None
    return descriptor_contents


def descriptor_copy(ctx, descriptor_files, app_dir):
    for f in descriptor_files:
        if ctx.Contains(f):
            with open(os.path.join(app_dir, f), 'w') as w:
                w.write(ctx.GetFile(f))


def gen_tmp_dir(dirr):
    tmp_dir = tempfile.mkdtemp()
    dir_name = os.path.join(tmp_dir, dirr)
    os.mkdir(dir_name)
    return dir_name


def creation_time(image):
    logging.info(image.config_file())
    cfg = json.loads(image.config_file())
    return cfg.get('created')


def timestamp_to_time(dt_str):
    dt = dt_str.rstrip('Z')
    return datetime.datetime.strptime(dt, "%Y-%m-%dT%H:%M:%S")


def generate_overrides(set_path):
    env = {
        'VIRTUAL_ENV': '/env',
    }
    if set_path:
        env['PATH'] = '/env/bin:$PATH'
    overrides_dct = {
        'created': str(datetime.date.today()) + 'T00:00:00Z',
        'env': env
    }
    return overrides_dct
