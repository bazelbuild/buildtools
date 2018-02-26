workspace(name = "com_github_bazelbuild_buildtools")

# 0.10.0
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.0/rules_go-0.10.0.tar.gz",
    sha256 = "53c8222c6eab05dd49c40184c361493705d4234e60c42c4cd13ab4898da4c6be",
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
    sha256 = "255e1199c0876b9a8cc02d5ea569b6cfe1901d30428355817b7606ddecb04c15",
    strip_prefix = "bazel-0.8.0",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.8.0.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.8.0.tar.gz",
    ],
)
