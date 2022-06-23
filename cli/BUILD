load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "clear.go",
        "common.go",
        "config.go",
        "configure-docker.go",
        "dockerHelper.go",
        "gcr-login.go",
        "gcr-logout.go",
        "version.go",
    ],
    importpath = "github.com/GoogleCloudPlatform/docker-credential-gcr/cli",
    visibility = ["//visibility:public"],
    deps = [
        "//auth:go_default_library",
        "//config:go_default_library",
        "//credhelper:go_default_library",
        "//store:go_default_library",
        "@com_github_docker_cli//cli/config:go_default_library",
        "@com_github_docker_cli//cli/config/configfile:go_default_library",
        "@com_github_docker_docker_credential_helpers//credentials:go_default_library",
        "@com_github_google_subcommands//:go_default_library",
    ],
)
