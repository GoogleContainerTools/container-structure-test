Docker Container Functional Tests
========================================

This code builds an image that serves as a framework to run functional tests on
a target image. The image under tests runs as a docker container inside of this
image, which itself runs as a docker container on a host machine in the cloud
(when run through a
[Google Cloud Container Build](https://cloud.google.com/container-builder/docs/overview)
).

# How Use This Image

This image can be first built using the included `cloudbuild.yaml`. The
following will build a `functional_test` GCR image in your project.

``` shell
gcloud container builds submit --config cloudbuild.yaml .
```

Then, you can run another container build using this image to execute the test.
An example is given under `examples/simple/cloudbuild.yaml`, which assumes
the `functional_test` image is in the same project, and will run the tests
specified in `test.yaml`. You can run this example as followed:

``` shell
cd examples/simple
gcloud container builds submit --config cloudbuild.yaml .
```

Essentially, `test.yaml` is specified as an argument to `functional_test`:

``` yaml
steps:
- name: gcr.io/$PROJECT_ID/functional_test
  args: [--test_spec, test.yaml]
```

# Test Specifications

The tests are specified by a YAML or JSON file. Currently, this image can only
execute tests from one file at a time.

## Setup

The setup section of the file specifies commands to setup a running container
to test. These commands run once for all tests.

This section can spin up multiple containers should the test subject need
dependencies.

For example, the following starts a container named `some-redis` running a redis
server and waits 5 seconds to make sure the server runs.

``` yaml
setup:
- command: [docker, run, --name, some-redis, -d, launcher.gcr.io/google/redis3]
- command: [sleep, 5s]
```

## Teardown

The teardown section of the file specifies zero or more commands to clean up
after the tests. These commands run once for all tests.

It is recommended for this section to remove the started containers to avoid
interferences with the other functional tests in the same build session.

For example, the following removes the container started in the example above:

``` yaml
teardown:
- command: [docker, stop, some-redis]
- command: [docker, rm, some-redis]
```

## Target

This specifies the name of the container started by the setup to run tests
against.

For example, we want to target the container named `some-redis` in the examples
above:

``` yaml
target: some-redis
```

## Tests

The tests section specifies zero or more tests. The tests are run in the order
they are specified. Each test can have side effects that subsequent tests depend
upon.

Each test has the following attributes:

| Name | Description |
|---|---|
| name | Gives the test a descriptive name. Optional. |
| command | A command, as an array of `[command, arg, arg...]` to run. Required. |
| expect | Defines the test's expectations. Optional. |

If a test is run just for a side effect, you can leave `expect` attribute empty.

`expect` has the following a attributes:

| Name | Description |
|---|---|
| stdout | Specifies expectations on the STDOUT. |
| stderr | Specifies expectations on the STDERR. |

Expectations on STDOUT and STDERR can use zero or more of the following attributes:

| Name | Description |
|---|---|
| equals | String. The output is trimmed before comparing. |
| exactly | String. The output is compared verbatim. |
| matches | String. The output must _partially_ match the regex defined. |
| mustBeEmpty | Boolean. If true, no output is expected. |

For example:

``` yaml
tests:
- name: Can set a new key
  command: [redis-cli, set, mykey, test-value]
  expect:
    stdout:
      equals: OK
    stderr:
      mustBeEmpty: true
- name: Can get a previously set key
  command: [redis-cli, get, mykey]
  expect:
    stdout:
      equals: test-value
    stderr:
      mustBeEmpty: true
```

## Substitutions

Test commands can use substitutions, which are specified in the test spec as
shell variable substitutions `${VAR_NAME}` or `$VAR_NAME`. The values of the
variables are specified in the command line argument `--vars VAR_NAME=value`.
This arugment can be specified multiple times.

# Development

`examples/build-and-test/cloudbuild.yaml` conveniently builds this image and run
the test in one build session. This is suitable for the development of this
image. Use the following command, from the main directory, to run a test build:

``` shell
gcloud container builds submit --config examples/build-and-test/cloudbuild.yaml .
```
