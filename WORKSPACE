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

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@com_github_bazelbuild_buildtools//buildifier:deps.bzl", "buildifier_dependencies")

gazelle_dependencies()

go_rules_dependencies()

go_register_toolchains()

buildifier_dependencies()

go_repository(
    name = "skylark_syntax",
    commit = "a5f7082aabed29c0e429c722292c66ec8ecf9591",
    importpath = "github.com/google/skylark",
)
