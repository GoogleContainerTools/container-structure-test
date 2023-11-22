"Exposes container-structure-test as a Bazel rule"

load("//bazel:container_structure_test.bzl", "lib")

container_structure_test = rule(
    implementation = lib.implementation,
    attrs = lib.attrs,
    doc = """\
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
""",
    test = True,
    toolchains = [
        "@aspect_bazel_lib//lib:jq_toolchain_type",
        "@bazel_tools//tools/sh:toolchain_type",
        "@container_structure_test//bazel:structure_test_toolchain_type",
    ],
)
