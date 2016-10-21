load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load(":build_defs.bzl", "genfile_check_test", "go_yacc")

go_yacc(
    src = "parse.y",
    out = "parse.y.baz.go",
)

# parse.y.go is checked in to satisfy the Go community
# https://github.com/bazelbuild/buildifier/issues/14
# this test ensures it doesn't get stale.
genfile_check_test(
    src = "parse.y.go",
    gen = "parse.y.baz.go",
)

go_library(
    name = "go_default_library",
    srcs = [
        "lex.go",
        "parse.y.baz.go",  # keep
        "print.go",
        "quote.go",
        "rewrite.go",
        "rule.go",
        "syntax.go",
        "tables.go",
        "walk.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = [
        "parse_test.go",
        "print_test.go",
        "quote_test.go",
    ],
    data = glob(["testdata/*"]),
    library = ":go_default_library",
)
