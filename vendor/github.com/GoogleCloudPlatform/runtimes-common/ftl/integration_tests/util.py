"""A script to generate a cloudbuild yaml."""

import os

# Add directories for new tests here.
TEST_DIRS = [
    'gcp_build_test', 'packages_test', 'packages_lock_test',
    'destination_test', 'npmrc_test'
]

_ST_IMAGE = ('gcr.io/gcp-runtimes/structure-test:'
             '6195641f5a5a14c63c7945262066270842150ddb')

INITIAL_CLOUDBUILD_YAML = {
    'steps': [
        # We need to chmod in some cases for permissions.
        {
            'name': 'ubuntu',
            'args': ['chmod', 'a+rx', '-R', '/workspace'],
            'id': 'chmod',
        }
    ]
}


def run_test_steps(builder_name, full_name, directory, args):
    return [
        # First build the image
        {
            'name': 'bazel/ftl:%s' % builder_name,
            'args': args,
            'id': 'build-image-%s' % full_name,
        },
        # Then pull it from the registry
        {
            'name': 'gcr.io/cloud-builders/docker',
            'args': ['pull', full_name],
            'id': 'pull-image-%s' % full_name,
        },
        # Then test it.
        {
            'name':
            _ST_IMAGE,
            'args': [
                '/go_default_test', '-image', full_name,
                os.path.join(directory, 'structure_test.yaml')
            ],
            'id':
            'test-image%s' % full_name
        }
    ]
