package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//python:image.bzl",
    "py_image",
)

py_binary(
    name = "php_benchmark",
    srcs = [
        "main.py",
        "//ftl/benchmark:benchmark_lib",
    ],
    data = ["//ftl:php_builder.par"],
    main = "main.py",
    deps = [
        "//ftl:ftl_lib",
        "@containerregistry",
    ],
)

load("@base_images_docker//dockerfile_build:dockerfile_build.bzl", "dockerfile_build")

dockerfile_build(
    name = "benchmark_base",
    base = "//ftl:php_builder_base",
    dockerfile = "//ftl/benchmark:Dockerfile",
)

py_image(
    name = "php_benchmark_image",
    srcs = [
        "main.py",
        "//ftl/benchmark:benchmark_lib",
    ],
    base = ":benchmark_base",
    main = "main.py",
    deps = [
        "//ftl:ftl_lib",
        "@containerregistry",
    ],
)
