# Copyright 2017 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_build",
)

py_library(
    name = "containerregistry_mock_lib",
    srcs = glob([
        "*.py",
    ]),
    deps = [
        "@containerregistry",
        "@mock",
    ],
)

docker_build(
    name = "test",
    base = "@distroless_base//image",
)

py_test(
    name = "mock_registry_tests",
    srcs = ["mock_registry_tests.py"],
    data = [":test.tar"],
    deps = [
        "@containerregistry",
        "@mock",
    ],
)
