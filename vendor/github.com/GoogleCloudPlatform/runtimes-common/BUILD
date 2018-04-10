# Exclude rewriting docgen/scripts path
# gazelle:exclude docgen/scripts

# Runtimes Common: A collections of scripts for cloud languages runtimes team
# to manage continuous integration for silver languages.

# This code is compiled into a docker image that is publicly available. If you're
# interested in using this code, use the docker image, not the source.
package(default_visibility = ["//visibility:private"])

licenses(["notice"])  # Apache 2.0

exports_files(["LICENSE"])

load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

go_prefix("github.com/GoogleCloudPlatform/runtimes-common")

load("@bazel_gazelle//:def.bzl", "gazelle")

gazelle(
    name = "gazelle",
    build_tags = ["go1.7"],
    external = "vendored",
    prefix = "github.com/GoogleCloudPlatform/runtimes-common",
)
