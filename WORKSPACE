workspace(name = "com_github_bazelbuild_buildtools")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f87fa87475ea107b3c69196f39c82b7bbf58fe27c62a338684c20ca17d1d8613",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.16.2/rules_go-0.16.2.tar.gz",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "6e875ab4b6bf64a38c352887760f21203ab054676d9c1b274963907e0768740d",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.15.0/bazel-gazelle-0.15.0.tar.gz",
)

http_archive(
    name = "bazel_skylib",
    sha256 = "b5f6abe419da897b7901f90cbab08af958b97a8f3575b0d3dd062ac7ce78541f",
    strip_prefix = "bazel-skylib-0.5.0",
    url = "https://github.com/bazelbuild/bazel-skylib/archive/0.5.0.tar.gz",
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
    sha256 = "dd07fb88a3f4c9bb68416eb277bfbea20c982a9f4bd6525368d4e4beea55cb57",
    strip_prefix = "bazel-0.19.2",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.19.2.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.19.2.tar.gz",
    ],
)

go_repository(
    name = "skylark_syntax",
    commit = "a5f7082aabed29c0e429c722292c66ec8ecf9591",
    importpath = "github.com/google/skylark",
)
