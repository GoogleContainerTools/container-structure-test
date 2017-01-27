Container Build Structure Tests
===============================

This code builds an image which serves as a framework to run structure-based tests on a target image as part of a CI/CD build flow. These tests can be run before pushing an image to GCR, or post-push as part of a continuous system. The image under tests runs as a docker container **inside** of this image, which itself runs as a docker container on a host machine in the cloud (when run through a [Google Cloud Container Build](https://cloud.google.com/container-builder/docs/overview)).

To use this test image with any cloudbuild, add the following build step to the **end** your container build config (cloudbuild.yaml or cloudbuild.json):

              name: gcr.io/gcp-runtimes/structure_test
              args:
                  - <your_target_image>

It's **very important that this step appears at the end of your build** (or at least after the image itself it assembled by Docker); without a built image, there will be nothing to test, and your build will fail!

Tests within this framework are specified through a JSON or YAML config file, by default called `structure_test.json` (this can be specified through a `--config` flag argument to the build step). This file will be copied into the workspace of the structure test image and loaded in by the test driver, which will execute the tests in order. Within this config file, three distinct types of tests can be written:

- Command Tests (testing output/error of a specific command issued)
- File Existence Tests (making sure a file is, or isn't, present in the file system of the image)
- File Content Tests (making sure files in the file system of the image contain, or do not contain, specific contents)

## Command Tests
Command tests ensure that certain commands run properly on top of the shell of the target image. Regexes can be used to check for expected or excluded strings in both stdout and stderr. Additionally, any number of flags can be passed to the argument as normal.

#### Supported Fields:

- Name (string, **required**): The name of the test
- Setup ([][]string, *optional*): A list of commands (each with optional flags) to run before the actual command under test.
- Teardown ([][]string, *optional*): A list of commands (each with optional flags) to run after the actual command under test.
- Command ([]string, **required**): The command to run, along with the flags to pass to it.
- Expected Output ([]string, *optional*): List of regexes that should match the stdout from running the command.
- Excluded Output ([]string, *optional*): List of regexes that should **not** match the stdout from running the command.
- Expected Error ([]string, *optional*): List of regexes that should match the stderr from running the command.
- Excluded Error ([]string, *optional*): List of regexes that should **not** match the stderr from running the command.
- Exit Code (int, *optional*): Exit code that the command should exit with.

Example:
```json
"commandTests": [
        {
                "name": "apt-get upgrade",
                "command": ["apt-get", "-qqs", "upgrade"],
                "excludedOutput": [".*Inst.*Security.* | .*Security.*Inst.*"],
                "excludedError": [".*Inst.*Security.* | .*Security.*Inst.*"]
        },
        {
                "name": "Custom Node Version",
                "setup": [["install_node", "v5.9.0"]],
                "teardown": [["install_node", "v6.9.1"]],
                "command": ["node", "-v"],
                "expectedOutput": ["v5.9.0\n"],
                "exitCode": 0
        }
]
```

```yaml
commandTests:
- name:  'apt-get'
  command: ['apt-get', 'help']
  expectedError: ['.*Usage.*']
  excludedError: ['*FAIL.*']
```


## File Existence Tests
File existence tests check to make sure a specific file (or directory) exist within the file system of the image. No contents of the files or directories are checked. These tests can also be used to ensure a file or directory is **not** present in the file system.

#### Supported Fields:

- Name (string, **required**): The name of the test
- Path (string, **required**): Path to the file or directory under test
- IsDirectory (boolean, **required**): Whether or not the specified path is a directory (as opposed to a file)
- ShouldExist (boolean, **required**): Whether or not the specified file or directory should exist in the file system
- Permissions (string, *optional*): The expected Unix permission string (e.g.
  drwxrwxrwx) of the files or directory.

Example:
```json
"fileExistenceTests": [
        {
                "name": "Root",
                "path": "/",
                "isDirectory": true,
                "shouldExist": true,
                "permissions": "-rw-r--r--"
        },
        {
                "name": "Fake file",
                "path": "/foo/bar",
                "isDirectory": false,
                "shouldExist": false
        }
]
```

Example:
```yaml
fileExistenceTests:
- name: 'Root'
  path: '/'
  isDirectory: true
  shouldExist: true
  permissions: '-rw-r--r--'
```

## File Content Tests
File content tests open a file on the file system and check its contents. These tests assume the specified file **is a file**, and that it **exists** (if unsure about either or these criteria, see the above **File Existence Tests** section). Regexes can again be used to check for expected or excluded content in the specified file.

#### Supported Fields:

- Name (string, **required**): The name of the test
- Path (string, **required**): Path to the file under test
- ExpectedContents (string[], *optional*): List of regexes that should match the contents of the file
- ExcludedContents (string[], *optional*): List of regexes that should **not** match the contents of the file

Example:
```json
"fileContentTests": [
        {
                "name": "Debian Sources",
                "path": "/etc/apt/sources.list",
                "expectedContents": [
                        ".*httpredir\\.debian\\.org.*"
                ],
                "excludedContents": [
                        ".*gce_debian_mirror.*"
                ]
        }
]
```

Example:
```yaml
fileContentTests:
- name: 'Debian Sources'
  path: '/etc/apt/sources.list'
  expectedContents: ['.*httpredir\\.debian\\.org.*']
  excludedContents: ['.*gce_debian_mirror.*']
```

## License Tests
License tests check a list of copyright files and makes sure all licenses are
allowed at Google. By default it will look at where Debian lists all copyright
files, but can also look at an arbitrary list of files.

#### Supported Fields:

- Debian (bool, **required**): If the image is based on Debian, check where
  Debian lists all licenses.
- Files (string[], *optional*): A list of other files to check.

Example:
```json
"licenseTests": [
      {
            "debian": true,
            "files": ["/foo/bar", "/baz/bat"]
      }
]
```

Example:
```yaml
licenseTests:
- debian: true
  files: ["/foo/bar", "/baz/bat"]
```

### Running Structure Tests Outside of Container Build
Structure tests can also be run outside of Cloud Container Build through a shell script, `ext_run.sh`. This allows the structure test framework to be used as normal presubmit tests in build systems like TravisCI. The only requirement to run is that the host machine has a working installation of Docker.

This script will retrieve the static structure_test binary from the published Docker image, volume mount it (along with all specified config files) into the container built from the image under test, and run the tests as normal. Supported arguments:

Sample fetch and run:

```shell
curl https://raw.githubusercontent.com/GoogleCloudPlatform/runtimes-common/master/structure_tests/ext_run.sh > ext_run.sh
bash ext_run.sh -i gcr.io/gcp-runtimes/check_if_tag_exists -c check_if_image_tag_exists/test_config.json
```

Flags:
- [--image, -i]: The image to be tested (e.g. gcr.io/gcp-runtimes/check_if_tag_exists)
- [--verbose, -v]: Boolean flag to show verbose logging/output from structure tests
- [--config, -c]: JSON config file defining the actual tests to be run (Note: any number of config files may be specified)
