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

mkdir -p test_dir/subdir
mkdir -p golden
INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > test_dir/build  # case doesn't matter
echo -e "$INPUT" > test_dir/test.bzl
echo -e "$INPUT" > test_dir/subdir/test.bzl
echo -e "$INPUT" > test.bzl  # outside the test_dir directory
echo -e "not valid +" > test_dir/foo.bar
cp test_dir/foo.bar golden/foo.bar

"$buildifier" < test_dir/build > stdout
"$buildifier" -r test_dir
"$buildifier" test.bzl
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

diff test_dir/build golden/BUILD.golden
diff test_dir/test.bzl golden/test.bzl.golden
diff test_dir/subdir/test.bzl golden/test.bzl.golden
diff test_dir/foo.bar golden/foo.bar
diff test.bzl golden/test.bzl.golden
diff stdout golden/test.bzl.golden
diff test_dir/test.bzl.out golden/test.bzl.golden

# Test run on a directory without -r
"$buildifier" test_dir || ret=$?
if [[ $ret -ne 3 ]]; then
  die "Directory without -r: expected buildifier to exit with 3, actual: $ret"
fi

# Test the linter

cat > test_dir/to_fix.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

cat > test_dir/fixed_golden.bzl <<EOF
a = b // c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_all.bzl <<EOF
a = b // c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_dict_cfg.bzl <<EOF
a = b / c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test_dir/fixed_golden_cfg.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test_dir/fix_report_golden <<EOF
test_dir/to_fix_tmp.bzl: applied fixes, 1 warnings left
fixed test_dir/to_fix_tmp.bzl
EOF

error_docstring="test_dir/to_fix_tmp.bzl:1: module-docstring: The file has no module docstring. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#module-docstring)"
error_integer="test_dir/to_fix_tmp.bzl:1: integer-division: The \"/\" operator for integer division is deprecated in favor of \"//\". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division)"
error_dict="test_dir/to_fix_tmp.bzl:2: unsorted-dict-items: Dictionary items are out of their lexicographical order. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#unsorted-dict-items)"
error_cfg="test_dir/to_fix_tmp.bzl:3: attr-cfg: cfg = \"data\" for attr definitions has no effect and should be removed. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#attr-cfg)"

test_lint () {
  ret=0
  cp test_dir/to_fix.bzl test_dir/to_fix_tmp.bzl
  echo "$4" > golden/error_golden

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

  # --lint=fix
  $buildifier --lint=fix $2 -v test_dir/to_fix_tmp.bzl 2> test_dir/fix_report || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: fix: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test_dir/to_fix_tmp.bzl $3 || die "$1: wrong file output for --lint=fix"
  diff test_dir/fix_report golden/fix_report_golden || die "$1: wrong console output for --lint=fix"
}

test_lint "default" "" "test_dir/fixed_golden.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_cfg" 1
test_lint "all" "--warnings=all" "test_dir/fixed_golden_all.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
test_lint "cfg" "--warnings=attr-cfg" "test_dir/fixed_golden_cfg.bzl" "$error_cfg" 0
test_lint "custom" "--warnings=-integer-division,+unsorted-dict-items" "test_dir/fixed_golden_dict_cfg.bzl" "$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
