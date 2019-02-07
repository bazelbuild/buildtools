set -e

die () {
  echo "$1" 1>&2
  exit 1
}

buildifier=$1
buildifier2=$2

mkdir test
INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > test/build  # case doesn't matter
echo -e "$INPUT" > test/test.bzl

"$buildifier" < test/build > stdout
"$buildifier" test/*
"$buildifier2" test/test.bzl > test/test.bzl.out

cat > test/BUILD.golden <<EOF
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
cat > test/test.bzl.golden <<EOF
load(":foo.bzl", "foo")

foo(tags = ["b", "a"], srcs = ["d", "c"])
EOF

diff test/build test/BUILD.golden
diff test/test.bzl test/test.bzl.golden
diff stdout test/BUILD.golden  # should use the build formatting mode by default

diff test/test.bzl.out test/test.bzl.golden

# Test the linter

cat > test/to_fix.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

cat > test/fixed_golden.bzl <<EOF
a = b // c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test/fixed_golden_all.bzl <<EOF
a = b // c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test/fixed_golden_dict_cfg.bzl <<EOF
a = b / c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > test/fixed_golden_cfg.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > test/fix_report_golden <<EOF
test/to_fix_tmp.bzl: applied fixes, 1 warnings left
fixed test/to_fix_tmp.bzl
EOF

error_docstring="test/to_fix_tmp.bzl:1: module-docstring: The file has no module docstring. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#module-docstring)"
error_integer="test/to_fix_tmp.bzl:1: integer-division: The \"/\" operator for integer division is deprecated in favor of \"//\". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division)"
error_dict="test/to_fix_tmp.bzl:2: unsorted-dict-items: Dictionary items are out of their lexicographical order. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#unsorted-dict-items)"
error_cfg="test/to_fix_tmp.bzl:3: attr-cfg: cfg = \"data\" for attr definitions has no effect and should be removed. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#attr-cfg)"

test_lint () {
  ret=0
  cp test/to_fix.bzl test/to_fix_tmp.bzl
  echo "$4" > test/error_golden

  cat > test/fix_report_golden <<EOF
test/to_fix_tmp.bzl: applied fixes, $5 warnings left
fixed test/to_fix_tmp.bzl
EOF

  $buildifier --lint=warn $2 test/to_fix_tmp.bzl 2> test/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test/error test/error_golden || die "$1: wrong console output for --lint=warn"

  $buildifier --lint=fix $2 -v test/to_fix_tmp.bzl 2> test/fix_report || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: fix: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff test/to_fix_tmp.bzl $3 || die "$1: wrong file output for --lint=fix"
  diff test/fix_report test/fix_report_golden || die "$1: wrong console output for --lint=fix"
}

test_lint "default" "" "test/fixed_golden.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_cfg" 1
test_lint "all" "--warnings=all" "test/fixed_golden_all.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
test_lint "cfg" "--warnings=attr-cfg" "test/fixed_golden_cfg.bzl" "$error_cfg" 0
test_lint "custom" "--warnings=-integer-division,+unsorted-dict-items" "test/fixed_golden_dict_cfg.bzl" "$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
