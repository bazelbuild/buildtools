workspace(name = "com_github_bazelbuild_buildifier")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "0691986e0754b38d33dd7c187a3e6bfd1fb0e5a1dfb2073fa97d02d255ee7ee2",
    strip_prefix = "rules_go-6fab60252e89cb603afce15d0d0321758895ffd2",
    urls = [
        "http://bazel-mirror.storage.googleapis.com/github.com/bazelbuild/rules_go/archive/6fab60252e89cb603afce15d0d0321758895ffd2.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/6fab60252e89cb603afce15d0d0321758895ffd2.tar.gz",
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
