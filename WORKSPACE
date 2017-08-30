workspace(name = "com_github_bazelbuild_buildtools")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "1e8e662ab93eca94beb6c690b8fd41347835e8ce0f3c4f71708af4b6673dd171",
    strip_prefix = "rules_go-2e319588571f20fdaaf83058b690abd32f596e89",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/rules_go/archive/2e319588571f20fdaaf83058b690abd32f596e89.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/2e319588571f20fdaaf83058b690abd32f596e89.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "new_go_repository")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_repositories")

go_repositories()

go_proto_repositories()

# used for build.proto
http_archive(
    name = "io_bazel",
    sha256 = "71e8b433b5d210867322336a2afcc8d11e832cb5db9e04100e3ac8bba2c9af96",
    strip_prefix = "bazel-0.5.4",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.5.4.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.5.4.tar.gz",
    ],
)

new_go_repository(
    name = "org_golang_x_tools",
    commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
    importpath = "golang.org/x/tools",
)
