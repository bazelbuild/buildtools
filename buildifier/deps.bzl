load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_dependencies():
    _maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "7363ae6721c1648017e23a200013510c9e71ca69f398d52886ee6af7f26af436",
        strip_prefix = "bazel-skylib-c00ef493869e2966d47508e8625aae723a4a3054",
        url = "https://github.com/bazelbuild/bazel-skylib/archive/c00ef493869e2966d47508e8625aae723a4a3054.tar.gz",  # 2018-12-06
    )
    _maybe(
        http_archive,
        name = "io_bazel",
        sha256 = "f59608e56b0b68fe9b18661ae3d10f6a61aaa5f70ed11f2db52e7bc6db516454",
        strip_prefix = "bazel-0.20.0",
        urls = [
            "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.20.0.tar.gz",
            "https://github.com/bazelbuild/bazel/archive/0.20.0.tar.gz",
        ],
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
