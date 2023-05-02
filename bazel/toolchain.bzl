"""This module implements the toolchain rule to locate the binary."""

StructureTestInfo = provider(
    doc = "Information about how to invoke the container-structure-test executable.",
    fields = {
        "binary": "Executable container-structure-test binary",
    },
)

def _structure_test_toolchain_impl(ctx):
    binary = ctx.executable.structure_test

    template_variables = platform_common.TemplateVariableInfo({
        "STRUCTURE_TEST_BIN": binary.path,
    })
    default = DefaultInfo(
        files = depset([binary]),
        runfiles = ctx.runfiles(files = [binary]),
    )
    st_info = StructureTestInfo(binary = binary)

    toolchain_info = platform_common.ToolchainInfo(
        st_info = st_info,
        template_variables = template_variables,
        default = default,
    )
    return [
        default,
        toolchain_info,
        template_variables,
    ]

structure_test_toolchain = rule(
    implementation = _structure_test_toolchain_impl,
    attrs = {
        "structure_test": attr.label(
            doc = "A hermetically downloaded structure_test executable for the target platform.",
            mandatory = True,
            executable = True,
            cfg = "exec",
            allow_single_file = True,
        ),
    },
    doc = "Defines a structure_test toolchain. See: https://docs.bazel.build/versions/main/toolchains.html#defining-toolchains.",
)
