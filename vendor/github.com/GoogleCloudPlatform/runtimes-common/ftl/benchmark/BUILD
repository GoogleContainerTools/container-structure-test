package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//python:image.bzl",
    "py_image",
)

exports_files([
    "Dockerfile",
])

py_library(
    name = "benchmark_lib",
    srcs = glob(["*.py"]),
)

py_test(
    name = "benchmark_test",
    srcs = [
        ":benchmark_lib",
        "//ftl:ftl_lib",
    ],
    data = [
        "//ftl/node/benchmark:node_benchmark",
        "//ftl/php/benchmark:php_benchmark",
        "//ftl/python/benchmark:python_benchmark",
    ],
    main = "benchmark_test.py",
)
