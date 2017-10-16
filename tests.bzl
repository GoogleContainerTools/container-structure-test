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

load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_build",
)

def _impl(ctx):

    if (not (ctx.attr.image_tar or ctx.executable.image) or
        (ctx.attr.image and ctx.executable.image)):
        fail('Please specify one of \'image\' or \'image_tar\'')
    st_binary = ctx.executable._structure_test.short_path
    config_location = ctx.file.config.short_path
    if ctx.executable.image:
        load_statement = ctx.executable.image.short_path
        image = ctx.attr.image
        runfiles = [ctx.executable.image] + image.files.to_list() + image.data_runfiles.files.to_list()
    else: # image_tar was set.
        load_statement = 'docker load -i %s' % ctx.file.image_tar.short_path
        image = ctx.attr.image_tar
        runfiles = [ctx.file.image_tar]

    image_name = "bazel/%s:%s" % (image.label.package, image.label.name)
    # Generate a shell script to execute structure_tests with the correct flags.
    test_contents = """\
#!/bin/bash
set -ex
# Execute the image load statement.
{0}

# Run the tests.
{1} \
  -image {2} \
  $(pwd)/{3}
""".format(load_statement, st_binary, image_name, config_location)
    ctx.file_action(
        output=ctx.outputs.executable,
        content=test_contents
    )

    return struct(runfiles=ctx.runfiles(files = [
        ctx.executable._structure_test,
        ctx.file.config] + 
        runfiles
        ),
    )

structure_test = rule(
    attrs = {
        "_structure_test": attr.label(
            default = Label("//structure_tests:go_default_test"),
            cfg = "target",
            allow_files = True,
            executable = True,
        ),
        "image": attr.label(
            executable = True,
            cfg = "target",
        ),
        "image_tar": attr.label(
            allow_files = [".tar"],
            single_file = True,
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

def structure_test_with_files(name, image, config, files):
  """A macro for including extra files inside an image before testing it."""
  child_image_name = "%s.child_image" % name
  docker_build(
      name = child_image_name,
      base = image,
      files = files,
  )

  structure_test(
      name = name,
      image = child_image_name,
      config = config,
  )
