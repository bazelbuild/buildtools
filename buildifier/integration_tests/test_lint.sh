#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/test_utils.sh"

init_buildifier "$1"
setup_test_env

# Tests linting warnings, fixes, and various lint modes.

cat > test_dir/to_fix.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 / 2
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

cat > test_dir/fixed_golden.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 // 2
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_all.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 // 2
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_dict_cfg.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 / 2
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_cfg.bzl <<EOF
load("//foo/bar/internal/baz:module.bzl", "b")

b()
a = 1 / 2
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fix_report_golden <<EOF
test_dir/to_fix_tmp.bzl: applied fixes, 2 warnings left
fixed test_dir/to_fix_tmp.bzl
EOF

error_bzl="test_dir/to_fix_tmp.bzl:1: bzl-visibility: Module \"//foo/bar/internal/baz:module.bzl\" can only be loaded from files located inside \"//foo/bar\", not from \"//test_dir/to_fix_tmp.bzl\". (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#bzl-visibility)"
error_docstring="test_dir/to_fix_tmp.bzl:1: module-docstring: The file has no module docstring."$'\n'"A module docstring is a string literal (not a comment) which should be the first statement of a file (it may follow comment lines). (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#module-docstring)"
error_integer="test_dir/to_fix_tmp.bzl:4: integer-division: The \"/\" operator for integer division is deprecated in favor of \"//\". (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#integer-division)"
error_dict="test_dir/to_fix_tmp.bzl:5: unsorted-dict-items: Dictionary items are out of their lexicographical order. (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#unsorted-dict-items)"
error_cfg="test_dir/to_fix_tmp.bzl:6: attr-cfg: cfg = \"data\" for attr definitions has no effect and should be removed. (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#attr-cfg)"

test_lint () {
  ret=0
  cp test_dir/to_fix.bzl test_dir/to_fix_tmp.bzl
  echo "$4" > golden/error_golden
  with_replaced_dir="${4//test_dir/another_test_dir}"
  echo "${with_replaced_dir//\/\/to_fix_tmp.bzl///another_test_dir/to_fix_tmp.bzl}" > golden/error_golden_another

  cat > golden/fix_report_golden <<EOF
test_dir/to_fix_tmp.bzl: applied fixes, $5 warnings left
fixed test_dir/to_fix_tmp.bzl
EOF

  # --lint=warn with --mode=check
  $buildifier --mode=check --lint=warn $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff -u test_dir/error golden/error_golden || die "$1: wrong console output for --mode=check --lint=warn"
  diff -u test_dir/to_fix.bzl test_dir/to_fix.bzl || die "$1: --mode=check --lint=warn shouldn't modify files"

  # --lint=warn
  $buildifier --lint=warn $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff -u test_dir/error golden/error_golden || die "$1: wrong console output for --lint=warn"

  # --lint=warn with --path
  $buildifier --lint=warn --path=another_test_dir/to_fix_tmp.bzl $2 test_dir/to_fix_tmp.bzl 2> test_dir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff -u test_dir/error golden/error_golden_another || die "$1: wrong console output for --lint=warn and --path"

  # --lint=fix
  $buildifier --lint=fix $2 -v test_dir/to_fix_tmp.bzl 2> test_dir/fix_report || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: fix: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff -u test_dir/to_fix_tmp.bzl $3 || die "$1: wrong file output for --lint=fix"
  diff -u test_dir/fix_report golden/fix_report_golden || die "$1: wrong console output for --lint=fix"
}

test_lint "default" "" "test_dir/fixed_golden.bzl" "$error_bzl"$'\n'"$error_docstring"$'\n'"$error_integer"$'\n'"$error_cfg" 2
test_lint "all" "--warnings=all" "test_dir/fixed_golden_all.bzl" "$error_bzl"$'\n'"$error_docstring"$'\n'"$error_integer"$'\n'"$error_dict"$'\n'"$error_cfg" 2
test_lint "cfg" "--warnings=attr-cfg" "test_dir/fixed_golden_cfg.bzl" "$error_cfg" 0
test_lint "custom" "--warnings=-bzl-visibility,-integer-division,+unsorted-dict-items" "test_dir/fixed_golden_dict_cfg.bzl" "$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
