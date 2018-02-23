load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mocks.go"],
    importpath = "github.com/GoogleCloudPlatform/docker-credential-gcr/mock/mock_store",
    visibility = ["//visibility:public"],
    deps = [
        "//store:go_default_library",
        "//vendor/github.com/docker/docker-credential-helpers/credentials:go_default_library",
        "//vendor/github.com/golang/mock/gomock:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
    ],
)
