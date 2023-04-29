"Implementation details for container_structure_test rule."

load("@aspect_bazel_lib//lib:paths.bzl", "BASH_RLOCATION_FUNCTION", "to_rlocation_path")
load("@aspect_bazel_lib//lib:windows_utils.bzl", "create_windows_native_launcher_script")

_attrs = {
    "image": attr.label(
        allow_single_file = True,
        doc = "Label of an oci_image or oci_tarball target.",
    ),
    "configs": attr.label_list(allow_files = True, mandatory = True),
    "driver": attr.string(
        default = "docker",
        # https://github.com/GoogleContainerTools/container-structure-test/blob/5e347b66fcd06325e3caac75ef7dc999f1a9b614/pkg/drivers/driver.go#L26-L28
        values = ["docker", "tar", "host"],
        doc = "See https://github.com/GoogleContainerTools/container-structure-test#running-file-tests-without-docker",
    ),
    "_runfiles": attr.label(default = "@bazel_tools//tools/bash/runfiles"),
    "_windows_constraint": attr.label(default = "@platforms//os:windows"),
}

CMD = """\
#!/usr/bin/env bash

{BASH_RLOCATION_FUNCTION}

readonly st=$(rlocation {st_path})
readonly yq=$(rlocation {yq_path})
readonly image=$(rlocation {image_path})

# When the image points to a folder, we can read the index.json file inside
if [[ -d "$image" ]]; then
  readonly DIGEST=$("$yq" eval '.manifests[0].digest | sub(":"; "-")' "${{image}}/index.json")
fi

exec "$st" test {fixed_args} $@
"""

def _structure_test_impl(ctx):
    fixed_args = ["--driver", ctx.attr.driver]
    test_bin = ctx.toolchains["@container_structure_test//bazel:structure_test_toolchain_type"].st_info.binary
    yq_bin = ctx.toolchains["@aspect_bazel_lib//lib:yq_toolchain_type"].yqinfo.bin

    image_path = to_rlocation_path(ctx, ctx.file.image)
    # Prefer to use a tarball if we are given one, as it works with more 'driver' types.
    if image_path.endswith(".tar"):
        fixed_args.extend(["--image", "$(rlocation %s)" % image_path])
    else:
        # https://github.com/GoogleContainerTools/container-structure-test/blob/5e347b66fcd06325e3caac75ef7dc999f1a9b614/cmd/container-structure-test/app/cmd/test.go#L110
        if ctx.attr.driver != "docker":
            fail("when the 'driver' attribute is not 'docker', then the image must be a .tar file")
        fixed_args.extend(["--image-from-oci-layout", "$(rlocation %s)" % image_path])
        fixed_args.extend(["--default-image-tag", "registry.structure_test.oci.local/image:$DIGEST"])

    for arg in ctx.files.configs:
        fixed_args.extend(["--config", "$(rlocation %s)" % to_rlocation_path(ctx, arg)])

    bash_launcher = ctx.actions.declare_file("%s.sh" % ctx.label.name)
    ctx.actions.write(
        bash_launcher,
        content = CMD.format(
            BASH_RLOCATION_FUNCTION = BASH_RLOCATION_FUNCTION,
            st_path = to_rlocation_path(ctx, test_bin),
            yq_path = to_rlocation_path(ctx, yq_bin),
            image_path = image_path,
            fixed_args = " ".join(fixed_args),
        ),
        is_executable = True,
    )

    is_windows = ctx.target_platform_has_constraint(ctx.attr._windows_constraint[platform_common.ConstraintValueInfo])
    launcher = create_windows_native_launcher_script(ctx, bash_launcher) if is_windows else bash_launcher

    runfiles = ctx.runfiles(
        files = ctx.files.image + ctx.files.configs + [
            bash_launcher, test_bin, yq_bin
        ]).merge(ctx.attr._runfiles.default_runfiles)

    return DefaultInfo(runfiles = runfiles, executable = launcher)

lib = struct(
    attrs = _attrs,
    implementation = _structure_test_impl,
)
