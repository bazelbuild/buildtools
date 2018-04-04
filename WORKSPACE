workspace(name = "com_github_bazelbuild_buildtools")

git_repository(
    name = "io_bazel_rules_go",
    commit = "9b3a85e62cc8c00d4b356bfb2035ca0657703187",
    remote = "https://github.com/bazelbuild/rules_go.git",
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
    sha256 = "e5321afb93e75cfb55f6f9c34d44f15230f8103677aa48a76ce3e868ee490d8e",
    strip_prefix = "bazel-0.11.1",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.11.1.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.11.1.tar.gz",
    ],
)

go_repository(
    name = "skylark_syntax",
    commit = "f11011f2887ba17f71cf974fc319dbb550a48ed5",
    importpath = "github.com/google/skylark",
)
