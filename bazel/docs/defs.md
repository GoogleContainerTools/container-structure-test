<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Exposes container-structure-test as a Bazel rule

<a id="container_structure_test"></a>

## container_structure_test

<pre>
load("@container_structure_test//:defs.bzl", "container_structure_test")

container_structure_test(<a href="#container_structure_test-name">name</a>, <a href="#container_structure_test-configs">configs</a>, <a href="#container_structure_test-driver">driver</a>, <a href="#container_structure_test-image">image</a>, <a href="#container_structure_test-platform">platform</a>)
</pre>

Tests a Docker- or OCI-format image.

By default, it relies on the container runtime already installed and running on the target.

By default, container-structure-test uses the socket available at `/var/run/docker.sock`.
If the installation creates the socket in a different path, use
`--test_env=DOCKER_HOST='unix://<path_to_sock>'`.

If the installation uses a remote Docker daemon and is protected by TLS, the following may be needed as well
`--test_env=DOCKER_TLS_VERIFY=1`
`--test_env=DOCKER_CERT_PATH=<path_to_certs>`.

To avoid putting this into the commandline or to instruct bazel to read it from terminal environment,
simply add `test --test_env=DOCKER_HOST` into the `.bazelrc` file.

Alternatively, use the `driver = "tar"` attribute to avoid the need for a container runtime, see
https://github.com/GoogleContainerTools/container-structure-test#running-file-tests-without-docker

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="container_structure_test-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="container_structure_test-configs"></a>configs |  -   | <a href="https://bazel.build/concepts/labels">List of labels</a> | required |  |
| <a id="container_structure_test-driver"></a>driver |  See https://github.com/GoogleContainerTools/container-structure-test#running-file-tests-without-docker   | String | optional |  `"docker"`  |
| <a id="container_structure_test-image"></a>image |  Label of an oci_image or oci_tarball target.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="container_structure_test-platform"></a>platform |  Set platform if host is multi-platform capable (default "linux/amd64")   | String | optional |  `"linux/amd64"`  |


