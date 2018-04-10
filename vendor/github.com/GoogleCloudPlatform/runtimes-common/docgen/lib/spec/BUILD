package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = ["spec.go"],
    importpath = "github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/spec",
    deps = [
        "//vendor/github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/proto:go_default_library",
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/github.com/golang/protobuf/jsonpb:go_default_library",
    ],
)

go_test(
    name = "spec_test",
    srcs = ["spec_test.go"],
    data = glob(["testdata/*"]),
    deps = [
        ":go_default_library",
        "@com_github_ghodss_yaml//:go_default_library",
        "@com_github_golang_protobuf//jsonpb:go_default_library",
        "@in_gopkg_yaml_v2//:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["spec_test.go"],
    data = glob(["testdata/**"]),
    importpath = "github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/spec_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/github.com/golang/protobuf/jsonpb:go_default_library",
    ],
)
