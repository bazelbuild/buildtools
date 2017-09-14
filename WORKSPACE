workspace(name = "com_github_bazelbuild_buildtools")

# 0.5.5
http_archive(
    name = "io_bazel_rules_go",
    strip_prefix = "rules_go-e1c4b58c05e4a6ab67392daf28f3d57e4902f581",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/rules_go/archive/e1c4b58c05e4a6ab67392daf28f3d57e4902f581.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/e1c4b58c05e4a6ab67392daf28f3d57e4902f581.tar.gz",
    ],
)

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_rules_dependencies",
    "go_register_toolchains",
    "go_repository",
)

go_rules_dependencies()

go_register_toolchains()

# used for build.proto
http_archive(
    name = "io_bazel",
    strip_prefix = "bazel-0.5.4",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.5.4.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.5.4.tar.gz",
    ],
)

go_repository(
    name = "org_golang_x_tools",
    commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
    importpath = "golang.org/x/tools",
)
