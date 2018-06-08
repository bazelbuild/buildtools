workspace(name = "com_github_bazelbuild_buildtools")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

git_repository(
    name = "io_bazel_rules_go",
    # 0.11.1
    commit = "12fa5fd88725c9033fc7c37ec0f04b64a9519f49",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "92a3c59734dad2ef85dc731dbcb2bc23c4568cded79d4b87ebccd787eb89e8d0",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.11.0/bazel-gazelle-0.11.0.tar.gz",
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
    sha256 = "e5321afb93e75cfb55f6f9c34d44f15230f8103677aa48a76ce3e868ee490d8e",
    strip_prefix = "bazel-0.11.1",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.11.1.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.11.1.tar.gz",
    ],
)

go_repository(
    name = "skylark_syntax",
    commit = "ede9b31f30c07f7081ae3c112b223d024c7f7a15",
    importpath = "github.com/google/skylark",
)
