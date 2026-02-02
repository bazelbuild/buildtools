#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/test_utils.sh"

init_buildifier "$1"
setup_test_env

# Tests JSON output format.

cat > test_dir/to_fix.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 / 2
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

mkdir test_dir/json
cp test_dir/to_fix.bzl test_dir/json

# just not formatted
cat > test_dir/json/to_fix_2.bzl <<EOF
a=b
EOF

# not formatted with rewrites
cat > test_dir/json/to_fix_3.bzl <<EOF
x = 0123
EOF

# formatted, no warnings
cat > test_dir/json/to_fix_4.bzl <<EOF
a = b
EOF

# not a starlark file
cat > test_dir/json/foo.bar <<EOF
this is not a starlark file
EOF

cat > golden/json_report_golden <<EOF
{
    "success": false,
    "files": [
        {
            "filename": "to_fix.bzl",
            "formatted": true,
            "valid": true,
            "warnings": [
                {
                    "start": {
                        "line": 1,
                        "column": 6
                    },
                    "end": {
                        "line": 1,
                        "column": 41
                    },
                    "category": "bzl-visibility",
                    "actionable": true,
                    "autoFixable": false,
                    "message": "Module \"//foo/bar/internal/baz:module.bzl\" can only be loaded from files located inside \"//foo/bar\", not from \"//test_dir/json/to_fix.bzl\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#bzl-visibility"
                },
                {
                    "start": {
                        "line": 4,
                        "column": 5
                    },
                    "end": {
                        "line": 4,
                        "column": 10
                    },
                    "category": "integer-division",
                    "actionable": true,
                    "autoFixable": true,
                    "message": "The \"/\" operator for integer division is deprecated in favor of \"//\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#integer-division"
                },
                {
                    "start": {
                        "line": 6,
                        "column": 15
                    },
                    "end": {
                        "line": 6,
                        "column": 27
                    },
                    "category": "attr-cfg",
                    "actionable": true,
                    "autoFixable": true,
                    "message": "cfg = \"data\" for attr definitions has no effect and should be removed.",
                    "url": "https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#attr-cfg"
                }
            ]
        },
        {
            "filename": "to_fix_2.bzl",
            "formatted": false,
            "valid": true,
            "warnings": []
        },
        {
            "filename": "to_fix_3.bzl",
            "formatted": false,
            "valid": true,
            "warnings": []
        },
        {
            "filename": "to_fix_4.bzl",
            "formatted": true,
            "valid": true,
            "warnings": []
        }
    ]
}
EOF

cat > golden/json_report_small_golden <<EOF
{
    "success": true,
    "files": [
        {
            "filename": "to_fix_4.bzl",
            "formatted": true,
            "valid": true,
            "warnings": []
        }
    ]
}
EOF

cat > golden/json_report_stdin_golden <<EOF
{
    "success": true,
    "files": [
        {
            "filename": "\u003cstdin\u003e",
            "formatted": true,
            "valid": true,
            "warnings": []
        }
    ]
}
EOF

cat > golden/json_report_invalid_file_golden <<EOF
{
    "success": false,
    "files": [
        {
            "filename": "to_fix_4.bzl",
            "formatted": true,
            "valid": true,
            "warnings": []
        },
        {
            "filename": "foo.bar",
            "formatted": false,
            "valid": false,
            "warnings": []
        }
    ]
}
EOF

cd test_dir/json

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v to_fix.bzl to_fix_2.bzl to_fix_3.bzl to_fix_4.bzl > json_report
diff -u json_report ../../golden/json_report_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with many files"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v to_fix_4.bzl > json_report
diff -u json_report ../../golden/json_report_small_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with a single file"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v < to_fix_4.bzl > json_report
diff -u json_report ../../golden/json_report_stdin_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with stdin"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v to_fix_4.bzl foo.bar > json_report
diff -u json_report ../../golden/json_report_invalid_file_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with an invalid file"

cd ../..
