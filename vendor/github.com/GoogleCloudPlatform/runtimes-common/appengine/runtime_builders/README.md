Runtime Builder
===============

This script takes a cloudbuild YAML config file, with all of it's build step images templated with tag names, resolves each of those tag names to a specific SHA256 digest, and uploads that pinned config file to GCS. The main purpose of this is to ensure that when a runtime builder is written, it is pinned to specific versions of each of its build steps to make a perfectly reproducible build.

Runtime maintainers will write their builds as a list of cloudbuild steps that the image will go through as it is being assembled and tested. Each of these steps is itself a Docker image, so in the cloudbuild config file the specific tag of the image to be used should be specified. In order to eliminate confusion, the tags must be templated in a specific format different from the way tags are normally specified, e.g.:

				name: gcr.io/gcp-runtimes/structure_test:${latest}

When this script is run, it will find and replace all templated tags in the config file with their corresponding digests. For example, the image specified above might be converted into something like:

				name: gcr.io/gcp-runtimes/structure_test@sha256:a3e9e8b880c5ca4322f2ea5f7fb90c16841f3da9072cb497e136ba2f44601e29

Once all of the tags in the file have been substituted out, the new file will be written to a bucket in Google Cloud Storage. This bucket will be accessible to builders and maintainers of the runtime images, and will be used anytime a new image needs to be built.