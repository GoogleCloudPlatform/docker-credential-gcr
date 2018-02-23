load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["file.go"],
    importpath = "github.com/docker/cli/cli/config/configfile",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/docker/docker/api/types:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
    ],
)
