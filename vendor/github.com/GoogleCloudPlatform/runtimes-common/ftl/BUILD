package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//python:image.bzl",
    "py_image",
)
load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_build",
)
load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_build",
)

py_library(
    name = "ftl_lib",
    srcs = glob([
        "__init__.py",
        "common/*.py",
    ]),
    deps = [
        "@containerregistry",
        "@httplib2",
    ],
)

py_library(
    name = "test_lib",
    srcs = ["common/test_util.py"],
    deps = [
        "//testing/lib:containerregistry_mock_lib",
        "@mock",
    ],
)

py_test(
    name = "context_test",
    srcs = ["common/context_test.py"],
    deps = [
        ":ftl_lib",
        "@mock",
    ],
)

py_test(
    name = "builder_test",
    srcs = ["common/builder_test.py"],
    deps = [
        ":ftl_lib",
        "@mock",
    ],
)

py_test(
    name = "cache_test",
    srcs = ["common/cache_test.py"],
    deps = [
        ":ftl_lib",
        "//testing/lib:containerregistry_mock_lib",
        "@mock",
    ],
)

py_test(
    name = "node_builder_test",
    srcs = ["node/builder_test.py"],
    data =
        glob(
            ["node/testdata/**"],
        ),
    main = "node/builder_test.py",
    deps = [
        ":node_lib",
        ":test_lib",
    ],
)

py_library(
    name = "node_lib",
    srcs = glob(["node/*.py"]),
    deps = [":ftl_lib"],
)

load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

pkg_tar(
    name = "default_creds",
    srcs = [":config.json"],
    mode = "0600",
    package_dir = "/root/.docker",
)

pkg_tar(
    name = "cred_helper",
    srcs = ["@docker_credential_gcr//:docker-credential-gcr"],
    mode = "0555",
    package_dir = "/usr/local/bin",
)

# The node base image sets a default CMD, which causes issues if you run our image with no CMD.
# So, reset it to [""] before building our py_image.
docker_build(
    name = "node_builder_base",
    base = "@node_base//image",
    cmd = [""],
    entrypoint = "",
    tars = [
        ":cred_helper.tar",
        ":default_creds.tar",
    ],
)

load("@subpar//:subpar.bzl", "par_binary")

par_binary(
    name = "node_builder",
    srcs = [":node_lib"],
    main = "node/main.py",
    deps = [
        "//testing/lib:containerregistry_mock_lib",
        "@containerregistry",
    ],
)

py_image(
    name = "node_builder_image",
    srcs = [":node_lib"],
    base = ":node_builder_base",
    main = "node/main.py",
    deps = [
        "@containerregistry",
    ],
)

py_library(
    name = "python_lib",
    srcs = glob(["python/*.py"]),
    deps = [":ftl_lib"],
)

py_test(
    name = "python_builder_test",
    srcs = ["python/builder_test.py"],
    data =
        glob(
            ["python/testdata/**"],
        ),
    main = "python/builder_test.py",
    deps = [
        ":python_lib",
        "//testing/lib:containerregistry_mock_lib",
        "@mock",
    ],
)

# The python base image sets a default CMD, which causes issues if you run our image with no CMD.
# So, reset it to [""] before building our py_image.
docker_build(
    name = "python_builder_base",
    base = "@python_base//image",
    cmd = [""],
    entrypoint = "",
    tars = [
        ":cred_helper.tar",
        ":default_creds.tar",
    ],
)

par_binary(
    name = "python_builder",
    srcs = [":python_lib"],
    main = "python/main.py",
    deps = ["@containerregistry"],
)

py_image(
    name = "python_builder_image",
    srcs = [":python_lib"],
    base = ":python_builder_base",
    main = "python/main.py",
    deps = [
        "@containerregistry",
    ],
)

# PHP

py_library(
    name = "php_lib",
    srcs = glob(["php/*.py"]),
    deps = [":ftl_lib"],
)

par_binary(
    name = "php_builder",
    srcs = [":php_lib"],
    main = "php/main.py",
    deps = [
        "//testing/lib:containerregistry_mock_lib",
        "@containerregistry",
    ],
)

py_test(
    name = "php_builder_test",
    srcs = ["php/builder_test.py"],
    data =
        glob(
            ["php/testdata/**"],
        ),
    main = "php/builder_test.py",
    deps = [
        ":php_lib",
        "//testing/lib:containerregistry_mock_lib",
        "@mock",
    ],
)

docker_build(
    name = "php_builder_base",
    base = "@php_base//image",
    cmd = [""],
    entrypoint = "",
    tars = [
        ":cred_helper.tar",
        ":default_creds.tar",
    ],
)

py_image(
    name = "php_builder_image",
    srcs = [":php_lib"],
    base = ":php_builder_base",
    main = "php/main.py",
    deps = [
        "@containerregistry",
    ],
)
