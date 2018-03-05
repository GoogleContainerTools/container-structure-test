Container Structure Tests
====================

The Container Structure Tests provide a powerful framework to validate the structure
of a container image. These tests can be used to check the output of commands
in an image, as well as verify metadata and contents of the filesystem.

Tests can be run either through a standalone binary, or through a Docker image.

## Installation
Download the latest binary release [here](https://storage.googleapis.com/container-structure-test/latest/container-structure-test),
or pull the image at `gcr.io/gcp-runtimes/container-structure-test`.
**Please note that at this time the binary is only compatible with Linux.**

## Setup
To use container structure tests to validate your containers, you need the following:
- The container structure test binary or docker image
- A container image to test against
- A test `.yaml` or `.json` file with user defined structure tests to run inside of the specified container image

Note that the test framework looks for the provided image in the local Docker
daemon (if it is not provided as a tar). The `-pull` flag can optionally
be provided to force a pull of a remote image before running the tests.

## Example Run
An example run of the test framework:
```shell
./structure-test -test.v -image gcr.io/google-appengine/python \
python_test_config.yaml
```
This command will run the tests on the Google App Engine Python image, with verbose logging,
using the `python_test_config.yaml` test config.

Tests within this framework are specified through a YAML or JSON config file,
which is provided to the test driver as the last positional argument of the
command. Multiple config files may be specified in a single test run. The
config file will be loaded in by the test driver, which will execute the tests
in order. Within this config file, four types of tests can be written:

- Command Tests (testing output/error of a specific command issued)
- File Existence Tests (making sure a file is, or isn't, present in the
file system of the image)
- File Content Tests (making sure files in the file system of the image
contain, or do not contain, specific contents)
- Metadata Test, *singular* (making sure certain container metadata is correct)

## Command Tests
Command tests ensure that certain commands run properly in the target image.
Regexes can be used to check for expected or excluded strings in both `stdout`
and `stderr`. Additionally, any number of flags can be passed to the argument
as normal.

#### Supported Fields:

This is the current schema version (v2.0.0).

- Name (`string`, **required**): The name of the test
- Setup (`[][]string`, *optional*): A list of commands
(each with optional flags) to run before the actual command under test.
- Teardown (`[][]string`, *optional*): A list of commands
(each with optional flags) to run after the actual command under test.
- Command (`string`, **required**): The command to run in the test.
- Args (`[]string`, *optional*): The arguments to pass to the command.
- EnvVars (`[]EnvVar`, *optional*): A list of environment variables to set for
the individual test. See the **Environment Variables** section for more info.
- Expected Output (`[]string`, *optional*): List of regexes that should
match the stdout from running the command.
- Excluded Output (`[]string`, *optional*): List of regexes that should **not**
match the stdout from running the command.
- Expected Error (`[]string`, *optional*): List of regexes that should
match the stderr from running the command.
- Excluded Error (`[]string`, *optional*): List of regexes that should **not**
match the stderr from running the command.
- Exit Code (`int`, *optional*): Exit code that the command should exit with.

Example:
```yaml
commandTests:
  - name: "gunicorn flask"
    setup: [["virtualenv", "/env"], ["pip", "install", "gunicorn", "flask"]]
    command: "which"
    args: ["gunicorn"]
    expectedOutput: ["/env/bin/gunicorn"]
- name:  "apt-get upgrade"
  command: "apt-get"
  args: ["-qqs", "upgrade"]
  excludedOutput: [".*Inst.*Security.* | .*Security.*Inst.*"]
  excludedError: [".*Inst.*Security.* | .*Security.*Inst.*"]
```

### Image Entrypoint

To avoid unexpected behavior and output when running commands in the
containers, **all entrypoints are overwritten by default.** If your
entrypoint is necessary for the structure of your container, use the
`setup` field to call any scripts or commands manually before running
the tests.

```yaml
commandTests:
  ...
  setup: [["my_image_entrypoint.sh"]]
  ...
```

### Intermediate Artifacts
Each command test run creates either a container (with the `docker` driver) or
tar artifact (with the `tar` driver). By default, these are deleted after the
test run finishes, but the `-save` flag can optionally be passed to keep
these around. This would normally be used for debugging purposes.


## File Existence Tests
File existence tests check to make sure a specific file (or directory) exist
within the file system of the image. No contents of the files or directories
are checked. These tests can also be used to ensure a file or directory is
**not** present in the file system.

#### Supported Fields:

- Name (`string`, **required**): The name of the test
- Path (`string`, **required**): Path to the file or directory under test
- ShouldExist (`boolean`, **required**): Whether or not the specified file or
directory should exist in the file system
- Permissions (`string`, *optional*): The expected Unix permission string (e.g.
  drwxrwxrwx) of the files or directory.

Example:
```yaml
fileExistenceTests:
- name: 'Root'
  path: '/'
  shouldExist: true
  permissions: '-rw-r--r--'
```

## File Content Tests
File content tests open a file on the file system and check its contents.
These tests assume the specified file **is a file**, and that it **exists**
(if unsure about either or these criteria, see the above
**File Existence Tests** section). Regexes can again be used to check for
expected or excluded content in the specified file.

#### Supported Fields:

- Name (`string`, **required**): The name of the test
- Path (`string`, **required**): Path to the file under test
- ExpectedContents (`string[]`, *optional*): List of regexes that
should match the contents of the file
- ExcludedContents (`string[]`, *optional*): List of regexes that
should **not** match the contents of the file

Example:
```yaml
fileContentTests:
- name: 'Debian Sources'
  path: '/etc/apt/sources.list'
  expectedContents: ['.*httpredir\\.debian\\.org.*']
  excludedContents: ['.*gce_debian_mirror.*']
```

## Metadata Test
The Metadata test ensures the container is configured correctly. All
of these checks are optional.

#### Supported Fields:

- Env (`[]EnvVar`): A list of environment variable key/value pairs that should be set
in the container.
- Entrypoint (`[]string`): The entrypoint of the container
- Cmd (`[]string`): The CMD specified in the container.
- Exposed Ports (`[]string`): The ports exposed in the container.
- Volumes (`[]string`): The volumes exposed in the container.
- Workdir (`string`): The default working directory of the container.

Example:
```yaml
metadataTest:
  env:
    - key: foo
      value: baz
  exposedPorts: ["8080", "2345"]
  volumes: ["/test"]
  entrypoint: []
  cmd: ["/bin/bash"]
  workdir: ["/app"]
```

## License Tests
License tests check a list of copyright files and makes sure all licenses are
allowed at Google. By default it will look at where Debian lists all copyright
files, but can also look at an arbitrary list of files.

#### Supported Fields:

- Debian (`bool`, **required**): If the image is based on Debian, check where
  Debian lists all licenses.
- Files (`string[]`, *optional*): A list of other files to check.

Example:
```yaml
licenseTests:
- debian: true
  files: ["/foo/bar", "/baz/bat"]
```

### Environment Variables
A list of environment variables can optionally be specified as part of the
test setup. They can either be set up globally (for all test runs), or
test-local as part of individual command test runs (see the **Command Tests**
section above). Each environment variable is specified as a key-value pair.
Unix-style environment variable substitution is supported.

To specify, add a section like this to your config:

```yaml
globalEnvVars:
  - key: "VIRTUAL_ENV"
    value: "/env"
  - key: "PATH"
    value: "/env/bin:$PATH"
```

## Running File Tests On [Google Cloud Container Builder](https://cloud.google.com/container-builder/docs)

This tool is released as a builder image, tagged as
`gcr.io/gcp-runtimes/container-structure-test`, so you can specify tests in your
`cloudbuild.yaml`:

```yaml

steps:
# Build an image.
- name: 'gcr.io/cloud-builders/docker'
  args: ['build', '-t', 'gcr.io/$PROJECT_ID/image', '.']
# Test the image.
- name: 'gcr.io/gcp-runtimes/container-structure-test'
  args: ['-image', 'gcr.io/$PROJECT_ID/image', 'test_config.yaml']

# Push the image.
images: ['gcr.io/$PROJECT_ID/image']
```

## Running File Tests Without Docker

Container images can be represented in multiple formats, and the Docker image
is just one of them. At their core, images are just a series of layers, each
of which is a tarball, and so can be interacted with without a working Docker
daemon. While running command tests currently requires a functioning Docker
daemon on the host machine, File Existence/Content tests do not. This can be
particularly useful when dealing with images which have been `docker export`ed
or saved in a different image format than the Docker format. To run tests
without using a Docker daemon, a user can specify a different "driver" to use
in the tests, with the `-driver` flag.

An example test run with a different driver looks like:
```shell
./structure-test -driver tar -image gcr.io/google-appengine/python \
python_test_config.yaml
```

The currently supported drivers in the framework are:
- `docker`: the default driver.
Supports all tests, and uses the Docker daemon on the host to run them.
- `tar`: a tar driver, which converts an image to a single tarball before
interacting with it. Does *not* support command tests.


### Running Structure Tests Through Bazel
Structure tests can also be run through `bazel`.
To do so, load the rule and its dependencies in your `WORKSPACE`:
```BUILD
git_repository(
    name = "io_bazel_rules_docker",
    commit = "8aeab63328a82fdb8e8eb12f677a4e5ce6b183b1",
    remote = "https://github.com/bazelbuild/rules_docker.git",
)

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "repositories",
)
repositories()


load(
    "@io_bazel_rules_docker//contrib:test.bzl",
    "container_test",
)

# io_bazel_rules_go is the dependency of container_test rules.
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.9.0/rules_go-0.9.0.tar.gz",
    sha256 = "4d8d6244320dd751590f9100cf39fd7a4b75cd901e1f3ffdfd6f048328883695",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

```

and then include the rule definition in your `BUILD` file:

```BUILD
load("@io_bazel_rules_docker//contrib:test.bzl", "container_test")
```

Then, create a `container_test` rule, passing in your image and config
file as parameters:

```BUILD
container_build(
    name = "hello",
    base = "//java:java8",
    cmd = ["/HelloJava_deploy.jar"],
    files = [":HelloJava_deploy.jar"],
)


container_test(
    name = "hello_test",
    configs = ["testdata/hello.yaml"],
    image = ":hello",
)
```

See this [example repo](https://github.com/nkubala/structure-test-examples) for a full working example.
