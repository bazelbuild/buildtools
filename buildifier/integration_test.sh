#!/bin/bash

# --- begin runfiles.bash initialization ---
# Copy-pasted from Bazel's Bash runfiles library (tools/bash/runfiles/runfiles.bash).
set -euo pipefail
if [[ ! -d "${RUNFILES_DIR:-/dev/null}" && ! -f "${RUNFILES_MANIFEST_FILE:-/dev/null}" ]]; then
  if [[ -f "$0.runfiles_manifest" ]]; then
    export RUNFILES_MANIFEST_FILE="$0.runfiles_manifest"
  elif [[ -f "$0.runfiles/MANIFEST" ]]; then
    export RUNFILES_MANIFEST_FILE="$0.runfiles/MANIFEST"
  elif [[ -f "$0.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    export RUNFILES_DIR="$0.runfiles"
  fi
fi
if [[ -f "${RUNFILES_DIR:-/dev/null}/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
  source "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "${RUNFILES_MANIFEST_FILE:-/dev/null}" ]]; then
  source "$(grep -m1 "^bazel_tools/tools/bash/runfiles/runfiles.bash " \
            "$RUNFILES_MANIFEST_FILE" | cut -d ' ' -f 2-)"
else
  echo >&2 "ERROR: cannot find @bazel_tools//tools/bash/runfiles:runfiles.bash"
  exit 1
fi
# --- end runfiles.bash initialization ---

die () {
  echo "$1" 1>&2
  exit 1
}

[[ "$1" =~ external/* ]] && buildifier="${{1#external/}}" || buildifier="$TEST_WORKSPACE/$1"
[[ "$2" =~ external/* ]] && buildifier2="${{2#external/}}" || buildifier2="$TEST_WORKSPACE/$2"
buildifier="$(rlocation "$buildifier")"
buildifier2="$(rlocation "$buildifier2")"

touch WORKSPACE.bazel
rm -r test_dir || true
mkdir -p test_dir/subdir
mkdir -p golden
INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > test_dir/BUILD
echo -e "$INPUT" > test_dir/test.bzl
echo -e "$INPUT" > test_dir/subdir/test.bzl
echo -e "$INPUT" > test_dir/subdir/build  # lowercase, should be ignored by -r
echo -e "$INPUT" > test.bzl  # outside the test_dir directory
echo -e "$INPUT" > test2.bzl  # outside the test_dir directory
echo -e "not valid +" > test_dir/foo.bar
mkdir test_dir/workspace  # name of a starlark file, but a directory
mkdir test_dir/.git  # contents should be ignored
echo -e "a+b" > test_dir/.git/git.bzl

cp test_dir/foo.bar golden/foo.bar
cp test_dir/subdir/build golden/build
cp test_dir/.git/git.bzl golden/git.bzl

"$buildifier" < test_dir/BUILD > stdout
"$buildifier" -r test_dir
"$buildifier" test.bzl
"$buildifier" --path=foo.bzl test2.bzl
"$buildifier2" test_dir/test.bzl > test_dir/test.bzl.out

cat > golden/BUILD.golden <<EOF
load(":foo.bzl", "foo")

foo(
    srcs = [
        "c",
        "d",
    ],
    tags = [
        "a",
        "b",
    ],
)
EOF
cat > golden/test.bzl.golden <<EOF
load(":foo.bzl", "foo")

foo(tags = ["b", "a"], srcs = ["d", "c"])
EOF

diff test_dir/BUILD golden/BUILD.golden
diff test_dir/test.bzl golden/test.bzl.golden
diff test_dir/subdir/test.bzl golden/test.bzl.golden
diff test_dir/subdir/build golden/build
diff test_dir/foo.bar golden/foo.bar
diff test.bzl golden/test.bzl.golden
diff test2.bzl golden/test.bzl.golden
diff stdout golden/test.bzl.golden
diff test_dir/test.bzl.out golden/test.bzl.golden
diff test_dir/.git/git.bzl golden/git.bzl

# Test run on a directory without -r
"$buildifier" test_dir || ret=$?
if [[ $ret -ne 3 ]]; then
  die "Directory without -r: expected buildifier to exit with 3, actual: $ret"
fi

# Test the linter

cat > test_dir/to_fix.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

cat > test_dir/fixed_golden.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

a = b // c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_all.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

a = b // c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_dict_cfg.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

a = b / c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_cfg.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fix_report_golden <<EOF
test_dir/to_fix_tmp.bzl: applied fixes, 2 warnings left
fixed test_dir/to_fix_tmp.bzl
EOF

error_bzl="test_dir/to_fix_tmp.bzl:1: bzl-visibility: Module \"//foo/bar/internal/baz:module.bzl\" can only be loaded from files located inside \"//foo/bar\", not from \"//test_dir/to_fix_tmp.bzl\". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#bzl-visibility)"
error_docstring="test_dir/to_fix_tmp.bzl:1: module-docstring: The file has no module docstring."$'\n'"A module docstring is a string literal (not a comment) which should be the first statement of a file (it may follow comment lines). (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#module-docstring)"
error_integer="test_dir/to_fix_tmp.bzl:3: integer-division: The \"/\" operator for integer division is deprecated in favor of \"//\". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division)"
error_dict="test_dir/to_fix_tmp.bzl:4: unsorted-dict-items: Dictionary items are out of their lexicographical order. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#unsorted-dict-items)"
error_cfg="test_dir/to_fix_tmp.bzl:5: attr-cfg: cfg = \"data\" for attr definitions has no effect and should be removed. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#attr-cfg)"

test_lint () {
  ret=0
  cp test_dir/to_fix.bzl test_dir/to_fix_tmp.bzl
  echo "$4" > golden/error_golden
  echo "${4//test_dir/another_test_dir}" > golden/error_golden_another

  cat > golden/fix_report_golden <<EOF
test_dir/to_fix_tmp.bzl: applied fixes, $5 warnings left
fixed test_dir/to_fix_tmp.bzl
EOF

  # --lint=warn with --mode=check
  $buildifier --mode=check --lint=warn $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test_dir/error golden/error_golden || die "$1: wrong console output for --mode=check --lint=warn"
  diff test_dir/to_fix.bzl test_dir/to_fix.bzl || die "$1: --mode=check --lint=warn shouldn't modify files"

  # --lint=warn
  $buildifier --lint=warn $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test_dir/error golden/error_golden || die "$1: wrong console output for --lint=warn"

  # --lint=warn with --path
  $buildifier --lint=warn --path=another_test_dir/to_fix_tmp.bzl $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test_dir/error golden/error_golden_another || die "$1: wrong console output for --lint=warn and --path"

  # --lint=fix
  $buildifier --lint=fix $2 -v test_dir/to_fix_tmp.bzl 2> test_dir/fix_report || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: fix: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test_dir/to_fix_tmp.bzl $3 || die "$1: wrong file output for --lint=fix"
  diff test_dir/fix_report golden/fix_report_golden || die "$1: wrong console output for --lint=fix"
}

test_lint "default" "" "test_dir/fixed_golden.bzl" "$error_bzl"$'\n'"$error_docstring"$'\n'"$error_integer"$'\n'"$error_cfg" 2
test_lint "all" "--warnings=all" "test_dir/fixed_golden_all.bzl" "$error_bzl"$'\n'"$error_docstring"$'\n'"$error_integer"$'\n'"$error_dict"$'\n'"$error_cfg" 2
test_lint "cfg" "--warnings=attr-cfg" "test_dir/fixed_golden_cfg.bzl" "$error_cfg" 0
test_lint "custom" "--warnings=-bzl-visibility,-integer-division,+unsorted-dict-items" "test_dir/fixed_golden_dict_cfg.bzl" "$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1

# Test --format=json

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
                    "message": "Module \"//foo/bar/internal/baz:module.bzl\" can only be loaded from files located inside \"//foo/bar\", not from \"//test_dir/json/to_fix.bzl\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#bzl-visibility"
                },
                {
                    "start": {
                        "line": 3,
                        "column": 5
                    },
                    "end": {
                        "line": 3,
                        "column": 10
                    },
                    "category": "integer-division",
                    "actionable": true,
                    "message": "The \"/\" operator for integer division is deprecated in favor of \"//\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division"
                },
                {
                    "start": {
                        "line": 5,
                        "column": 15
                    },
                    "end": {
                        "line": 5,
                        "column": 27
                    },
                    "category": "attr-cfg",
                    "actionable": true,
                    "message": "cfg = \"data\" for attr definitions has no effect and should be removed.",
                    "url": "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#attr-cfg"
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
diff json_report ../../golden/json_report_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with many files"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v to_fix_4.bzl > json_report
diff json_report ../../golden/json_report_small_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with a single file"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v < to_fix_4.bzl > json_report
diff json_report ../../golden/json_report_stdin_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with stdin"

$buildifier --mode=check --format=json --lint=warn --warnings=-module-docstring -v to_fix_4.bzl foo.bar > json_report
diff json_report ../../golden/json_report_invalid_file_golden || die "$1: wrong console output for --mode=check --format=json --lint=warn with an invalid file"

cd ../..

# Test the multifile functionality

mkdir multifile
cd multifile

cat > lib.bzl <<EOF
def foo():
  """
  This is a function foo.

  Please use it in favor of the
  deprecated function bar
  """
  pass

def bar():
  """
  This is a function bar.

  Deprecated:
    please use foo instead.
  """
  pass
EOF

touch WORKSPACE

cat > BUILD <<EOF
load(":lib.bzl", "foo", "bar")
load(":nonexistent.bzl", "foo2", "bar2")
EOF

cat > report_golden <<EOF
BUILD:1: deprecated-function: The function "bar" defined in "//lib.bzl" is deprecated. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#deprecated-function)
EOF

$buildifier --lint=warn --warnings=deprecated-function BUILD 2> report || ret=$?
diff report_golden report || die "$1: wrong console output for multifile warnings (WORKSPACE exists)"
