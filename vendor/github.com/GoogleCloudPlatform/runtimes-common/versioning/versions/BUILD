package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = ["versions.go"],
    importpath = "github.com/GoogleCloudPlatform/runtimes-common/versioning/versions",
    deps = [
        "//vendor/gopkg.in/yaml.v2:go_default_library",
    ],
)
