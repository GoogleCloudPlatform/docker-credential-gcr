load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mocks.go"],
    importpath = "github.com/GoogleCloudPlatform/docker-credential-gcr/v2/mock/mock_cmd",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/mock/gomock:go_default_library"],
)
