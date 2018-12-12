load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_dependencies():
    _maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "3b61715da37bc552cba875351e0c79ae150450d4cf3844b54b8c03cd2d0f481b",
        strip_prefix = "bazel-skylib-d7c5518fa061ae18a20d00b14082705d3d2d885d",
        url = "https://github.com/bazelbuild/bazel-skylib/archive/d7c5518fa061ae18a20d00b14082705d3d2d885d.tar.gz",  # 2018-11-21
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
