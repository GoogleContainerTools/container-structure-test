"""A script to generate a cloudbuild yaml."""

import os
import yaml
import argparse

_TEST_TEMPLATE = '/workspace/ftl/%s/testdata'

_BASE_MAP = {
    "node": 'gcr.io/gae-runtimes/nodejs8_app_builder:latest',
    "php": 'gcr.io/gae-runtimes/php72_app_builder:latest',
    "python": 'gcr.io/google-appengine/python:latest',
}

parser = argparse.ArgumentParser(
    description='Generate cloudbuild yaml for FTL cache test.')

parser.add_argument(
    '--runtime',
    dest='runtime',
    action='store',
    choices=['node', 'php', 'python'],
    default=None,
    required=True,
    help='flag to select the runtime for the cache test')


def main():
    args = parser.parse_args()
    app_dir = 'packages_test'
    path = 'gcr.io/ftl-node-test/%s/cache/%s' % (args.runtime, app_dir)
    name = path + ':latest'
    cloudbuild_yaml = {
        'steps': [
            # We need to chmod in some cases for permissions.
            {
                'name': 'ubuntu',
                'args': ['chmod', 'a+rx', '-R', '/workspace']
            },
            # Build the runtime builder par file
            {
                'name': 'gcr.io/cloud-builders/bazel',
                'args': ['build', 'ftl:%s_builder.par' % args.runtime]
            },
            # Run the cache test
            {
                'name':
                'gcr.io/cloud-builders/bazel',
                'args': [
                    'run',
                    '//ftl/%s/cached:%s_cached' % (args.runtime, args.runtime),
                    '--', '--base', _BASE_MAP[args.runtime], '--name', name,
                    '--directory',
                    os.path.join(_TEST_TEMPLATE % args.runtime + app_dir)
                ]
            },
        ]
    }

    print yaml.dump(cloudbuild_yaml)


if __name__ == "__main__":
    main()
