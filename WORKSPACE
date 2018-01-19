http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.9.0/rules_go-0.9.0.tar.gz",
    sha256 = "4d8d6244320dd751590f9100cf39fd7a4b75cd901e1f3ffdfd6f048328883695",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
go_rules_dependencies()
go_register_toolchains()

# Import direct Go dependencies.
go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    commit = "5ccada7d0a7ba9aeb5d3aca8d3501b4c2a509fec", # HEAD @ Jan 19, 2018
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    commit = "b28fcf2b08a19742b43084fb40ab78ac6c3d8067", # HEAD @ Jan 19, 2018
)

# Import transitive Go dependencies.

# golang.org/x/oauth2/google depends on cloud.google.com/go/compute/metadata
go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    commit = "767c40d6a2e058483c25fa193e963a22da17236d", # v0.18.0
)