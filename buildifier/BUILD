load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_test")

go_binary(
    name = "buildifier",
    srcs = [
        "buildifier.go",
    ],
    deps = [
        "//core:go_default_library",
        "//differ:go_default_library",
    ],
)
