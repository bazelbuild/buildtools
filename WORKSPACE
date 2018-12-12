workspace(name = "com_github_bazelbuild_buildtools")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "278c669f9dd472a1687263e89647397b8e54b588f0228cb57071a2048a049a4d",
    strip_prefix = "rules_go-01e5a9f8483167962eddd167f7689408bdeb4e76",
    # 0.16.3
    url = "https://github.com/bazelbuild/rules_go/archive/01e5a9f8483167962eddd167f7689408bdeb4e76.tar.gz",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "f490124dd4b97c136cb8565f3aeefc2f2c1736afda7728b9b227b2b8aeadc88c",
    strip_prefix = "bazel-gazelle-44ce230b3399a5d4472198740358fcd825b0c3c9",
    url = "https://github.com/bazelbuild/bazel-gazelle/archive/44ce230b3399a5d4472198740358fcd825b0c3c9.tar.gz",  # 2018-12-10
)

http_archive(
    name = "bazel_skylib",
    sha256 = "7363ae6721c1648017e23a200013510c9e71ca69f398d52886ee6af7f26af436",
    strip_prefix = "bazel-skylib-c00ef493869e2966d47508e8625aae723a4a3054",
    url = "https://github.com/bazelbuild/bazel-skylib/archive/c00ef493869e2966d47508e8625aae723a4a3054.tar.gz",  # 2018-12-06
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
    sha256 = "f59608e56b0b68fe9b18661ae3d10f6a61aaa5f70ed11f2db52e7bc6db516454",
    strip_prefix = "bazel-0.20.0",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.20.0.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.20.0.tar.gz",
    ],
)

go_repository(
    name = "skylark_syntax",
    commit = "a5f7082aabed29c0e429c722292c66ec8ecf9591",
    importpath = "github.com/google/skylark",
)
