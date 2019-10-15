load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["homedir.go"],
    importmap = "github.com/GoogleCloudPlatform/docker-credential-gcr/vendor/github.com/docker/docker/pkg/homedir",
    importpath = "github.com/docker/docker/pkg/homedir",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/opencontainers/runc/libcontainer/user:go_default_library"],
)
