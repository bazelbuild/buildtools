workspace(name = "com_github_bazelbuild_buildtools")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "c1f52b8789218bb1542ed362c4f7de7052abcf254d865d96fb7ba6d44bc15ee3",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.12.0/rules_go-0.12.0.tar.gz",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "ddedc7aaeb61f2654d7d7d4fd7940052ea992ccdb031b8f9797ed143ac7e8d43",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.12.0/bazel-gazelle-0.12.0.tar.gz",
)

http_archive(
    name = "bazel_skylib",
    sha256 = "57e8737fbfa2eaee76b86dd8c1184251720c840cd9abe5c3f1566d331cdf7d65",
    strip_prefix = "bazel-skylib-0.4.0",
    url = "https://github.com/bazelbuild/bazel-skylib/archive/0.4.0.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

gazelle_dependencies()

go_rules_dependencies()

go_register_toolchains()

buildifier_dependencies()

# used for build.proto
http_archive(
    name = "io_bazel",
    sha256 = "66135f877d0cc075b683474c50b1f7c3e2749bf0a40e446f20392f44494fefff",
    strip_prefix = "bazel-0.12.0",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.12.0.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.12.0.tar.gz",
    ],
)

go_repository(
    name = "skylark_syntax",
    commit = "ede9b31f30c07f7081ae3c112b223d024c7f7a15",
    importpath = "github.com/google/skylark",
)
