"""A script to generate a cloudbuild yaml."""

import os
import yaml
import argparse

# Add directories for new tests here.
DEP_TESTS = ['small_app', 'medium_app', 'large_app']
APP_SIZE_TESTS = {
    'scratch_small': '5',
    'scratch_medium': '500',
    'scratch_large': '50000'
}
_DATA_DIR = '/workspace/ftl/node/benchmark/data/'
_NODE_BASE = 'gcr.io/gae-runtimes/nodejs8_app_builder:latest'

parser = argparse.ArgumentParser(
    description='Generate cloudbuild yaml for FTL benchmarking.')

parser.add_argument(
    '--iterations',
    action='store',
    type=int,
    default=5,
    help='Number of times to build the image.')

parser.add_argument(
    '--dep-test',
    dest='dep_test',
    action='store_true',
    default=False,
    help='Flag to enable to dependency test for the benchmark.')

parser.add_argument(
    '--app-size-test',
    dest='app_size_test',
    action='store_true',
    default=False,
    help='Flag to enable the app size test for the benchmark.')


def main():
    args = parser.parse_args()
    if not (args.dep_test and args.app_size):
        args.dep_test = True
        args.app_size = True

    cloudbuild_yaml = {
        'steps': [
            # We need to chmod in some cases for permissions.
            {
                'name': 'ubuntu',
                'args': ['chmod', 'a+rx', '-R', '/workspace']
            },
            # Build the FTL image from source and load it into the daemon.
            {
                'name':
                'gcr.io/cloud-builders/bazel',
                'args': [
                    'run', '//ftl/node/benchmark:node_benchmark_image', '--',
                    '--norun'
                ],
            },
            # Build the node builder par file
            {
                'name': 'gcr.io/cloud-builders/bazel',
                'args': ['build', 'ftl:node_builder.par']
            },
        ]
    }

    # Generate a set of steps for each test and add them.
    if args.dep_test:
        for app_dir in DEP_TESTS:
            cloudbuild_yaml['steps'] += dependency_test_step(
                app_dir, args.iterations)

    # Generate a set of steps for each test and add them.
    if args.app_size_test:
        for app_dir in APP_SIZE_TESTS:
            cloudbuild_yaml['steps'] += app_size_test_step(
                app_dir, args.iterations, APP_SIZE_TESTS[app_dir])

    print yaml.dump(cloudbuild_yaml)


def dependency_test_step(app_dir, iterations):
    name = 'gcr.io/ftl-node-test/benchmark_%s:latest' % app_dir
    return [
        # First build the image
        {
            'name':
            'bazel/ftl/node/benchmark:node_benchmark_image',
            'args': [
                '--base', _NODE_BASE, '--name', name, '--directory',
                os.path.join(_DATA_DIR + app_dir), '--description', app_dir,
                '--iterations',
                str(iterations)
            ]
        }
    ]


def app_size_test_step(app_dir, iterations, gen_files):
    name = 'gcr.io/ftl-node-test/benchmark_%s:latest' % app_dir
    return [
        # First build the image
        {
            'name':
            'bazel/ftl/node/benchmark:node_benchmark_image',
            'args': [
                '--base', _NODE_BASE, '--name', name, '--directory',
                os.path.join(_DATA_DIR + app_dir), '--description', app_dir,
                '--iterations',
                str(iterations), '--gen_files', gen_files
            ]
        }
    ]


if __name__ == "__main__":
    main()
