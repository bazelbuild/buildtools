workspace(name = "com_github_bazelbuild_buildifier")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "3c8a95b0c353acc260bd11f08918b156d40f41f7349433f19df7083e030c4fdb",
    strip_prefix = "rules_go-0.3.3",
    url = "https://github.com/bazelbuild/rules_go/archive/0.3.3.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "new_go_repository")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_repositories")

go_repositories()

go_proto_repositories()

# used for build.proto
http_archive(
    name = "io_bazel",
    sha256 = "8e4646898fa9298422e69767752680d34cbf21bcae01c401b11aa74fcdb0ef66",
    strip_prefix = "bazel-0.4.1",
    url = "https://github.com/bazelbuild/bazel/archive/0.4.1.tar.gz",
)

new_go_repository(
    name = "org_golang_x_tools",
    commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
    importpath = "golang.org/x/tools",
)
