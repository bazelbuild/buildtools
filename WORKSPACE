workspace(name = "com_github_bazelbuild_buildifier")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "b7759f01d29c075db177f688ffb4464aad2b8fbb7017f89a1d3819ce07f1d584",
    strip_prefix = "rules_go-0.3.1",
    url = "https://github.com/bazelbuild/rules_go/archive/0.3.1.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories")
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
