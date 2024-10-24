"""Repository rules for fetching pre-built container-test binaries"""
load("@aspect_bazel_lib//lib:repositories.bzl", "register_jq_toolchains")
load("//bazel:toolchains_repo.bzl", "PLATFORMS", "toolchains_repo")

# TODO(alexeagle): automate updates when new releases
# Run following command to make sure all checksums are correct.

# bazel build @structure_test_st_darwin_amd64//... @structure_test_st_darwin_arm64//... @structure_test_st_linux_arm64//... \
# @structure_test_st_linux_s390x//...  @structure_test_st_linux_amd64//... @structure_test_st_windows_amd64//...

_VERSION="v1.19.2"
_HASHES = {
    "darwin-amd64": "sha256-mkBKyy32nnivskrvlj0BbPnauhXnUckmBJJZzrDdKYU=",
    "darwin-arm64": "sha256-qdy4KDN143ar99X2kIwvXwyQO/NBqwMCZhmVJ65Pu3Y=",
    "linux-amd64": "sha256-u6b/uaz5Z0QMHMD+Iz2uM9ClCq4ULxLIWCYZAue/tAU=",
    "linux-arm64": "sha256-ebFieFZDy+X9ZbNAhdUi6KPKRyBe+bsizBaBqoimZ88=",
    "linux-ppc64le": "sha256-pkvTL+pb9/kxVxs0HBPf3CgTLD4z6mkdXlrZRU+iiIE=",
    "linux-s390x": "sha256-v2Nu34HxrGtJ1Op8qWMtoFi36CNW4hI83RRTVd7uq7s=",
    "windows-amd64.exe": "sha256-J9+eC2BlqXXhLIDyjMpFm10uFAQcV8HzPOuz76y1WbE=",
}

STRUCTURE_TEST_BUILD_TMPL = """\
# Generated by container/repositories.bzl
load("@container_structure_test//bazel:toolchain.bzl", "structure_test_toolchain")
structure_test_toolchain(
    name = "structure_test_toolchain", 
    structure_test = "structure_test"
)
"""

def _structure_test_repo_impl(repository_ctx):
    platform = repository_ctx.attr.platform.replace("_", "-")

    if platform.find("windows") != -1:
        platform = platform + ".exe"
    url = "https://github.com/GoogleContainerTools/container-structure-test/releases/download/{version}/container-structure-test-{platform}".format(
        version = _VERSION,
        platform = platform,
    )
    repository_ctx.download(
        url = url,
        output = "structure_test",
        integrity = _HASHES[platform],
        executable = True,
    )
    repository_ctx.file("BUILD.bazel", STRUCTURE_TEST_BUILD_TMPL)

structure_test_repositories = repository_rule(
    _structure_test_repo_impl,
    doc = "Fetch external tools needed for structure test toolchain",
    attrs = {
        "platform": attr.string(mandatory = True, values = PLATFORMS.keys()),
    },
)

# Wrapper macro around everything above, this is the primary API
def container_structure_test_register_toolchain(name, register = True):
    """Convenience macro for users which does typical setup.

    - create a repository for each built-in platform like "container_linux_amd64" -
      this repository is lazily fetched when node is needed for that platform.
    - create a repository exposing toolchains for each platform like "container_platforms"
    - register a toolchain pointing at each platform
    Users can avoid this macro and do these steps themselves, if they want more control.
    Args:
        name: base name for all created repos, like "container7"
        register: whether to call through to native.register_toolchains.
            Should be True for WORKSPACE users, but false when used under bzlmod extension
    """

    st_toolchain_name = "structure_test_toolchains"

    register_jq_toolchains(register = register)

    for platform in PLATFORMS.keys():
        structure_test_repositories(
            name = "{name}_st_{platform}".format(name = name, platform = platform),
            platform = platform,
        )

        if register:
            native.register_toolchains("@{}//:{}_toolchain".format(st_toolchain_name, platform))

    toolchains_repo(
        name = st_toolchain_name,
        toolchain_type = "@container_structure_test//bazel:structure_test_toolchain_type",
        # avoiding use of .format since {platform} is formatted by toolchains_repo for each platform.
        toolchain = "@%s_st_{platform}//:structure_test_toolchain" % name,
    )

def _st_extension_impl(_):
    container_structure_test_register_toolchain("structure_test", register = False)

extension = module_extension(
    implementation = _st_extension_impl,
)
