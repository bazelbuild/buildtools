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
        sha256 = "71b4294c18429fb730e223d1ceca135245fecda43191998d6afef39e1ba48e78",
        strip_prefix = "bazel-0.21.0",
        urls = [
            "https://github.com/bazelbuild/bazel/archive/0.21.0.tar.gz",
        ],
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
