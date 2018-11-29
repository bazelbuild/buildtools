workspace(name = "com_github_bazelbuild_buildtools")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "ed8b5e0ee6f8855b3bfb5bdb418ee76066f54ccc60fad53aff6e3ccd6d7610d0",
    strip_prefix = "rules_go-b5a862a50c434c36996cf273ea33240cf0d95640",
    url = "https://github.com/bazelbuild/rules_go/archive/b5a862a50c434c36996cf273ea33240cf0d95640.tar.gz",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "c3511bfefc5734df6388d18ffe7d31de266c2ce35c172e9da8bf7ab5ad6e44f5",
    strip_prefix = "bazel-gazelle-422ea009aca276245ac5152e5d598d1e2c3e2813",
    url = "https://github.com/bazelbuild/bazel-gazelle/archive/422ea009aca276245ac5152e5d598d1e2c3e2813.tar.gz", # 2018-11-28
)

http_archive(
    name = "bazel_skylib",
    sha256 = "3b61715da37bc552cba875351e0c79ae150450d4cf3844b54b8c03cd2d0f481b",
    strip_prefix = "bazel-skylib-d7c5518fa061ae18a20d00b14082705d3d2d885d",
    url = "https://github.com/bazelbuild/bazel-skylib/archive/d7c5518fa061ae18a20d00b14082705d3d2d885d.tar.gz",  # 2018-11-21
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
