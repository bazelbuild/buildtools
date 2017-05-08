workspace(name = "com_github_bazelbuild_buildtools")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f7e42a4c1f9f31abff9b2bdee6fe4db18bc373287b7e07a5b844446e561e67e2",
    strip_prefix = "rules_go-4c9a52aba0b59511c5646af88d2f93a9c0193647",
    urls = [
        "http://bazel-mirror.storage.googleapis.com/github.com/bazelbuild/rules_go/archive/4c9a52aba0b59511c5646af88d2f93a9c0193647.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/4c9a52aba0b59511c5646af88d2f93a9c0193647.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "new_go_repository")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_repositories")

go_repositories()

go_proto_repositories()

# used for build.proto
http_archive(
    name = "io_bazel",
    sha256 = "9fc591bc366dfcbdbc265c5daebbf30ca200ce3e69d14f17e5b20e2d487b2fee",
    strip_prefix = "bazel-0.4.5",
    urls = [
        "http://bazel-mirror.storage.googleapis.com/github.com/bazelbuild/bazel/archive/0.4.5.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.4.5.tar.gz",
    ],
)

new_go_repository(
    name = "org_golang_x_tools",
    commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
    importpath = "golang.org/x/tools",
)
