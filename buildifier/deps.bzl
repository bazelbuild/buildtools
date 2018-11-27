load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_dependencies():
    _maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "b5f6abe419da897b7901f90cbab08af958b97a8f3575b0d3dd062ac7ce78541f",
        strip_prefix = "bazel-skylib-0.5.0",
        urls = ["https://github.com/bazelbuild/bazel-skylib/archive/0.5.0.tar.gz"],
    )
    _maybe(
        http_archive,
        name = "io_bazel",
        sha256 = "dd07fb88a3f4c9bb68416eb277bfbea20c982a9f4bd6525368d4e4beea55cb57",
        strip_prefix = "bazel-0.19.2",
        urls = [
            "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.19.2.tar.gz",
            "https://github.com/bazelbuild/bazel/archive/0.19.2.tar.gz",
        ],
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
