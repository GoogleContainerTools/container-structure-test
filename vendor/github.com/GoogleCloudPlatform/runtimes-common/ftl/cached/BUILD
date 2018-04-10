package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//python:image.bzl",
    "py_image",
)

py_library(
    name = "cached_lib",
    srcs = glob(["*.py"]),
    deps = [
        "@containerregistry",
        "@mock",
    ],
)

py_test(
    name = "cached_test",
    srcs = [
        ":cached_lib",
        "//ftl:ftl_lib",
    ],
    data = [
        "//ftl/node/cached:node_cached",
        "//ftl/php/cached:php_cached",
        "//ftl/python/cached:python_cached",
    ],
    main = "cached_test.py",
)
