"""Repository rules for fetching pre-built container-test binaries"""

load("//bazel:toolchains_repo.bzl", "PLATFORMS", "toolchains_repo")

# TODO(alexeagle): automate updates when new releases
_VERSION = "v1.15.0"
_HASHES = {
    "darwin-amd64": "sha256-6QRBiuQh2NiZLB6TTz8uuXO6HdEU4Wy42Pv3xk/ihqY=",
    "linux-amd64": "sha256-6FiT0F6wcWECqvy2jrpBc0GpLlSbEmySQPK5HXzoPVo=",
    "linux-arm64": "sha256-4yVlbxwTgc4appxMgaVpXvQJ0Q/ZTu3sSy+sSLpI/kw=",
    "linux-ppc64le": "sha256-aFuBytG1AljuHff8BX0+xIZm58dpJsbv4K3zqj6phAw=",
    "linux-s390x": "sha256-1tB8LoN6qLu4JkItHLgIjo3O7U5y94o6kHbShyXtlO0=",
    "windows-amd64.exe": "sha256-ZzgSHeblwqDJaukjOQMrdxbZHM8gy1L/SzpVVOQglN0="
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

    # There is no arm64 version of structure test binary.
    # TODO: remove this after we start publishing one.
    if platform.find("darwin") != -1:
        platform = platform.replace("arm64", "amd64")
    elif platform.find("windows") != -1:
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
