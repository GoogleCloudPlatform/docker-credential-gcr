load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "lookup.go",
        "lookup_unix.go",
        "lookup_unsupported.go",
        "user.go",
    ],
    importmap = "github.com/GoogleCloudPlatform/docker-credential-gcr/vendor/github.com/opencontainers/runc/libcontainer/user",
    importpath = "github.com/opencontainers/runc/libcontainer/user",
    visibility = ["//visibility:public"],
)
