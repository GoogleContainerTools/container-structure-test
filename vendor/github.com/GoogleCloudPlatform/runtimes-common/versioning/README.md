# Description

Versioning tools for Dockerfile source repos.

- `dockerfiles` generates versionsed Dockerfiles from on a common template.
- `cloudbuild` generates a configuration file to build these Dockerfiles using
  [Google Container Builder](https://cloud.google.com/container-builder/docs/).

# Installation

- Install bazel: https://bazel.build
- Clone this repo:

``` shell
git clone https://github.com/GoogleCloudPlatform/runtimes-common.git
cd runtimes-common
```

- Build:

``` shell
bazel run //:gazelle
bazel build versioning/scripts/dockerfiles:dockerfiles
bazel build versioning/scripts/cloudbuild:cloudbuild
```

- Set the path to the built scripts:

``` shell
BAZEL_ARCH=linux_amd64_stripped
export PATH=$PATH:$PWD/bazel-bin/versioning/scripts/dockerfiles/${BAZEL_ARCH}/
export PATH=$PATH:$PWD/bazel-bin/versioning/scripts/cloudbuild/${BAZEL_ARCH}/
```

# Create `versions.yaml`

At root of the Dockerfile source repo, add a file called `versions.yaml`.
Follow the format defined in `versions.go`. See an example on
[github](https://github.com/GoogleCloudPlatform/mysql-docker).

Primary folders in the Dockerfile source repo:

- `templates` contains `Dockerfile.template`, which is a Go template for
  generating `Dockerfile`s.
- `tests` contains any tests that should be included in the generated cloud
  build configuration.
- Version folders as defined in `versions.yaml`. The `Dockerfile`s are
  generated into these folders. The folders should also contain all
  supporting files for each version, for example `docker-entrypoint.sh` files.

# Usage of `dockerfiles` command

``` shell
cd path/to/dockerfile/repo
dockerfiles
```

# Usage of `cloudbuild` command

``` shell
cd path/to/dockerfile/repo
cloudbuild > cloudbuild.yaml
```

You can use the generated `cloudbuild.yaml` file as followed:

``` shell
gcloud container builds submit --config=cloudbuild.yaml .
```
