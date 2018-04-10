runtimes-common
=============

This repository contains common tools and scripts for working with containers.

The primary use is for small tools used to build, test and distribute container images by GCP engineers, although other users might find them useful as well.

**If you're looking for the container structure tests, check out our [new dedicated repo](https://github.com/GoogleCloudPlatform/container-structure-test).**

## Projects

Projects in this repo are mainly organized in sub-directories.

See below for a list of the tools contained here.

* [FTL](./ftl/) - A set of tools for building language-runtime focused images "faster-than-light".
* [Integration Tests](./integration_tests/) - A set of tools for testing the functionality of language-based application containers on GCP.
* [reconciletags](./appengine/reconciletags/) - A source-based workflow tool for managing the tags on container images in GCR.
* [runtime_builders](./appengine/runtime_builders) - A tool for releasing sets of container images.
* [check_if_image_tag_exists](./appengine/check_if_image_tag_exists/) - A Container Builder step to help prevent overwriting images.
* [containerregistry testing](./testing/) - A Python library for testing containerregistry.

## Developing

You'll most likely need the `bazel` tool to build the code in this repository.
Follow these instructions to install and configure [bazel](https://bazel.build/).

We provide a pre-commit git hook for convenience.
Please install this before sending any commits via:

```shell
ln -s $(pwd)/hack/hooks/* .git/hooks/
