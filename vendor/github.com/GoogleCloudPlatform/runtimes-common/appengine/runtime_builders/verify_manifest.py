#!/usr/bin/python

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

import argparse
import logging
import sys
import yaml

import builder_util


def main():
    logging.getLogger().setLevel(logging.INFO)

    parser = argparse.ArgumentParser()
    parser.add_argument('--manifest', '-m',
                        help='path to runtime.yaml manifest',
                        required=True)
    args = parser.parse_args()

    verify_manifest(args.manifest)


def verify_manifest(manifest_file):
    """Verify that the provided runtime manifest is valid before publishing.

    Aliases are provided for runtime 'names' that can be included in users'
    application configuration files: this method ensures that all the aliases
    can resolve to actual builder files.

    All builders and aliases are turned into nodes in a graph, which is then
    traversed to be sure that all nodes lead down to a builder node.

    Example formatting of the manifest, showing both an 'alias' and
    an actual builder file:

    runtimes:
      java:
        target:
          runtime: java-openjdk
      java-openjdk:
        target:
          file: gs://runtimes/java-openjdk-1234.yaml
        deprecation:
          message: "openjdk is deprecated."
    """
    with open(manifest_file) as f:
        manifest = yaml.load(f)
        _verify_manifest_formatting(manifest)
        node_graph = _build_manifest_graph(manifest)
        _verify_manifest_graph(node_graph)


def _verify_manifest_formatting(manifest):
    try:
        if 'schema_version' not in manifest:
            logging.error('Manifest does not contain schema_version!')
            sys.exit(1)
        for key, val in manifest.get('runtimes').iteritems():
            file = val.get('target').get('file', '')
            if not file:
                continue
            if file.startswith('gs://'):
                logging.error('Builder file {0} should NOT be prefixed with '
                              'GCS bucket prefix or bucket name!'.format(file))
                sys.exit(1)
            file = builder_util.RUNTIME_BUCKET_PREFIX + file
            if not builder_util.file_exists(file):
                logging.error('File {0} not found in GCS!'
                              .format(file))
                sys.exit(1)

    except KeyError as ke:
        logging.error('Error encountered when verifying manifest: %s', ke)
        sys.exit(1)


def _verify_manifest_graph(node_graph):
    for _, node in node_graph.items():
        seen = set()
        child = node
        while True:
            seen.add(child)
            if not child.child:
                break
            elif child.child not in node_graph.keys():
                logging.error('Non-existent alias provided for {0}: {1}'
                              .format(child.name, child.child))
                sys.exit(1)
            child = node_graph[child.child]
            if child in seen:
                logging.error('Circular dependency found in manifest! '
                              'Check node {0}'.format(child))
                sys.exit(1)
        if not child.isBuilder:
            logging.error('No terminating builder for alias {0}'
                          .format(node.name))
            sys.exit(1)


def _build_manifest_graph(manifest):
    try:
        node_graph = {}
        for key, val in manifest.get('runtimes').iteritems():
            target = val.get('target', {})
            if not target:
                if 'deprecation' not in val:
                    logging.error('No target or deprecation specified for '
                                  'runtime: %s', key)
                    sys.exit(1)
                continue
            child = None
            isBuilder = 'file' in target.keys()
            if not isBuilder:
                child = target['runtime']
            node = node_graph.get(key, {})
            if not node:
                node_graph[key] = builder_util.Node(key, isBuilder, child)
        return node_graph
    except (KeyError, AttributeError) as ke:
        logging.error('Error encountered when verifying manifest: %s', ke)
        sys.exit(1)


if __name__ == '__main__':
    main()
