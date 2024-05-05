"""Repository rules for fetching pre-built container-test binaries"""
load("@aspect_bazel_lib//lib:repositories.bzl", "register_jq_toolchains")
load("//bazel:toolchains_repo.bzl", "PLATFORMS", "toolchains_repo")

# TODO(alexeagle): automate updates when new releases
# Run following command to make sure all checksums are correct.

# bazel build @structure_test_st_darwin_amd64//... @structure_test_st_darwin_arm64//... @structure_test_st_linux_arm64//... \
# @structure_test_st_linux_s390x//...  @structure_test_st_linux_amd64//... @structure_test_st_windows_amd64//...

_VERSION = "v1.18.0"
_HASHES = {
    "darwin-amd64": "sha256-5y1LUSMqGM6ObPvhxb8lX3hr1s9qmipEeeF22AuV/qM=",
    "darwin-arm64": "sha256-gyciqmGRpEJWJVerDk3DMsgQCwIbH96erVsplg79RXc=",
    "linux-amd64": "sha256-E3KUXKTtni6NzZCOURgkmrMFrBQwqAyuaKKgLdFeB1k=",
    "linux-arm64": "sha256-6ViiaUqdpYsypcH1KJ0so5I3FcBXD3AdAPYdeZe2RU0=",
    "linux-ppc64le": "sha256-WhrNx20HPr5YlOdAvvlXu4JLf99aR3VR7j24nS37cPs=",
    "linux-s390x": "sha256-pVMML4W7RyD/UnVcl3rZaGWFa6hQCEK1vxWYwr9TVAY=",
    "windows-amd64.exe": "sha256-DMMeVu5TH7WzRidInMdvMkxuYWpuuV8u4R7y9M1ObnU="
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
