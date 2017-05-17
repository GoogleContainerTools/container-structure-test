# Copyright 2017 Google Inc. All rights reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Rule for running structure tests."""

def _impl(ctx):
    ext_run_location = ctx.executable._structure_test.short_path
    tar_location = ctx.attr.image.files.to_list()[0].short_path
    config_location = ctx.file.config.short_path

    # docker_build rules always generate an image named 'bazel/$package:$name'.
    image_name = "bazel/%s:%s" % (ctx.attr.image.label.package, ctx.attr.image.label.name)

    # Generate a shell script to execute ext_run with the correct flags.
    test_contents = """\
#!/bin/bash
%s \
  -i %s \
  -t %s \
  -c %s""" % (ext_run_location, image_name, tar_location, config_location)
    ctx.file_action(
        output=ctx.outputs.executable,
        content=test_contents
    )

    return struct(runfiles=ctx.runfiles(files = [
        ctx.executable._structure_test,
        ctx.file.config] + 
        ctx.attr.image.files.to_list(),
    ))

structure_test = rule(
    attrs = {
        "_structure_test": attr.label(
            default = Label("//structure_tests:ext_run"),
            cfg = "target",
            allow_files = True,
            executable = True,
        ),
        "image": attr.label(
            mandatory = True,
        ),
        "config": attr.label(
            mandatory = True,
            allow_files = True,
            single_file = True,
        ),
    },
    executable = True,
    test = True,
    implementation = _impl,
)
