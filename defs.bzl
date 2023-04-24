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

To avoid putting this into the commandline or to instruct bazel to read it from terminal environment, 
simply add `test --test_env=DOCKER_HOST` into the `.bazelrc` file.

Alternatively, use the `driver = "tar"` attribute to avoid the need for a container runtime, see
https://github.com/GoogleContainerTools/container-structure-test#running-file-tests-without-docker
""",
    test = True,
    toolchains = [
        "@container_structure_test//bazel:st_toolchain_type",
        "@aspect_bazel_lib//lib:yq_toolchain_type",
    ],
)
