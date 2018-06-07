load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_dependencies():
    _maybe(
        http_archive,
        name = "bazel_skylib",
        strip_prefix = "bazel-skylib-0.4.0",
        sha256 = "57e8737fbfa2eaee76b86dd8c1184251720c840cd9abe5c3f1566d331cdf7d65",
        urls = ["https://github.com/bazelbuild/bazel-skylib/archive/0.4.0.tar.gz"],
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
