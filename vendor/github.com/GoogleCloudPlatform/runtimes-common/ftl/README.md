## FTL

"FTL" stands for "faster than light", and represents a strategy for constructing container images quickly.

In this context, the "speed of light" is considered to be the time taken to do a standard "docker build" followed by a "docker push" to a registry.

By constructing the container image layers cleverly and reproducibly, we can use the registry as a cache and speed up the build/push steps of many common language package managers.

This repository currently contains Cloud Build steps and binaries for Node.js, Python and PHP languages and package managers.

## Usage

The typical usage of an FTL binary is:

```shell
$ ftl.par --directory=$dir --base=$base --image=$img
```

This command can be read as "Build the source code in directory `$dir` into an image named `$img`, based on the image `$base`.

These binaries **do not depend on Docker**, and construct images directly in the registry.

As an example, we will demonstrate using the Node FTL builder to create a container for a node application from a node app's source code:
Assume we are deploying the node source for the app https://github.com/JustinBeckwith/cloudcats.  First download the node ftl.par file from https://storage.googleapis.com/gcp-container-tools/ftl/node/latest/ftl.par.  Then we run the ftl.par file pointing to our application:
```shell
$ ftl.par --directory=$HOME/cloudcats/web --base=gcr.io/google-appengine/nodejs:latest --image=gcr.i
o/my-project/cloudcats-node-app:latest
```

## Releases
Currently FTL is released in .par format for each supported runtime.  The latest release is v0.2.0, changelog [here](https://github.com/GoogleCloudPlatform/runtimes-common/blob/master/ftl/CHANGELOG.md)

### node

[v0.2.0](https://storage.googleapis.com/gcp-container-tools/ftl/node/node-v0.2.0/ftl.par)

[HEAD](https://storage.googleapis.com/gcp-container-tools/ftl/node/latest/ftl.par)

Specific version (based on git $COMMIT_SHA)
`https://storage.googleapis.com/gcp-container-tools/ftl/node/$COMMIT_SHA/ftl.par`

### python

[v0.2.0](https://storage.googleapis.com/gcp-container-tools/ftl/node/python-v0.2.0/ftl.par)

[HEAD](https://storage.googleapis.com/gcp-container-tools/ftl/python/latest/ftl.par)

Specific version (based on git $COMMIT_SHA)
`https://storage.googleapis.com/gcp-container-tools/ftl/python/$COMMIT_SHA/ftl.par`

### php
[v0.2.0](https://storage.googleapis.com/gcp-container-tools/ftl/php/php-v0.2.0/ftl.par)

[HEAD](https://storage.googleapis.com/gcp-container-tools/ftl/php/latest/ftl.par)

Specific version (based on git $COMMIT_SHA)
`https://storage.googleapis.com/gcp-container-tools/ftl/php/$COMMIT_SHA/ftl.par`

## Developing
To run the FTL integration tests, run the following command locally from the root directory:

```shell
python ftl/ftl_node_integration_tests_yaml.py | gcloud container builds submit --config /dev/fd/0 .
python ftl/ftl_php_integration_tests_yaml.py | gcloud container builds submit --config /dev/fd/0 .
gcloud container builds submit --config ftl/ftl_python_integration_tests.yaml .
```

## FTL Runtime Design Documents
[php](https://docs.google.com/document/d/1AB255g8N-J7IYEbhmiTRf29Ox1afEgs3df8GK4NBhrk/edit?usp=sharing)

[python](https://docs.google.com/document/d/15JOk_IFgaXwTSdge7XlxzzXDVVWKqvI5vJVhtHRNp_k/edit?usp=sharing)
