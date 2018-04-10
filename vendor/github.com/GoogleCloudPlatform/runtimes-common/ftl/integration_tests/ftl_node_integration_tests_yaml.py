"""A script to generate a cloudbuild yaml."""

import os
import yaml

import util

# Add directories for new tests here.
TEST_DIRS = [
    'gcp_build_test', 'packages_test', 'packages_lock_test',
    'destination_test', 'metadata_test', 'npmrc_test'
]

_TEST_DIR = '/workspace/ftl/node/testdata'
_NODE_BASE = 'gcr.io/gae-runtimes/nodejs8_app_builder:latest'


def main():

    cloudbuild_yaml = util.INITIAL_CLOUDBUILD_YAML
    cloudbuild_yaml['steps'].append(
        # Build the FTL image from source and load it into the daemon.
        {
            'name': 'gcr.io/cloud-builders/bazel',
            'args': ['run', '//ftl:node_builder_image', '--', '--norun'],
            'id': 'build-builder',
            'waitFor': [cloudbuild_yaml['steps'][0]['id']],
        }, )

    # Generate a set of steps for each test and add them.
    test_map = {}
    for test in TEST_DIRS:
        test_map[test] = [
            '--base', _NODE_BASE, '--name',
            'gcr.io/ftl-node-test/%s-image' % test, '--directory',
            os.path.join(_TEST_DIR, test), '--no-cache'
        ]
    test_map['destination_test'].extend(['--destination', '/alternative-app'])
    test_map['metadata_test'].extend(['--entrypoint', '/bin/echo'])
    test_map['metadata_test'].extend(['--exposed-ports', '8090,8091'])
    for test, args in test_map.iteritems():
        cloudbuild_yaml['steps'] += util.run_test_steps(
            'node_builder_image', 'gcr.io/ftl-node-test/%s-image' % test,
            os.path.join(_TEST_DIR, test), args)

    print yaml.dump(cloudbuild_yaml)


if __name__ == "__main__":
    main()
