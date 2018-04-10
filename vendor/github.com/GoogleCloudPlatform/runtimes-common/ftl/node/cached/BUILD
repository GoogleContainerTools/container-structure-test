package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_docker//python:image.bzl",
    "py_image",
)
load("@subpar//:subpar.bzl", "par_binary")

py_binary(
    name = "node_cached",
    srcs = [
        "main.py",
    ],
    data = ["//ftl:node_builder.par"],
    main = "main.py",
    deps = [
        "//ftl:ftl_lib",
        "//ftl/cached:cached_lib",
    ],
)
