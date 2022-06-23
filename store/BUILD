load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["store.go"],
    importpath = "github.com/GoogleCloudPlatform/docker-credential-gcr/store",
    visibility = ["//visibility:public"],
    deps = [
        "//config:go_default_library",
        "//util:go_default_library",
        "@com_github_docker_docker_credential_helpers//credentials:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
        "@org_golang_x_oauth2//google:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "store_integration_test.go",
        "store_unit_test.go",
    ],
    embed = [":go_default_library"],
    deps = select({
        "@io_bazel_rules_go//go/platform:aix": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:android": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:darwin": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:dragonfly": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:freebsd": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:illumos": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:ios": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:js": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:linux": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:netbsd": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:openbsd": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:plan9": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "@io_bazel_rules_go//go/platform:solaris": [
            "@org_golang_x_oauth2//:go_default_library",
        ],
        "//conditions:default": [],
    }),
)
