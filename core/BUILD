load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load(":build_defs.bzl", "go_yacc")

go_yacc(
    src = "parse.y",
    out = "parse.y.go",
)

go_library(
    name = "go_default_library",
    srcs = [
        "lex.go",
        "parse.y.go",
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
    name = "build_test",
    size = "small",
    srcs = [
        "parse_test.go",
        "print_test.go",
        "quote_test.go",
    ],
    data = glob(["testdata/*"]),
    library = ":go_default_library",
)
