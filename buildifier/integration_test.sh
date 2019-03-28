set -e

die () {
  echo "$1" 1>&2
  exit 1
}

buildifier=$1
buildifier2=$2

mkdir testdir
mkdir testdir/subdir
mkdir golden
INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > testdir/build  # case doesn't matter
echo -e "$INPUT" > testdir/test.bzl
echo -e "$INPUT" > testdir/subdir/test.bzl
echo -e "$INPUT" > test.bzl  # outside the testdir directory
echo -e "not valid +" > testdir/foo.bar
cp testdir/foo.bar golden/foo.bar

"$buildifier" < testdir/build > stdout
"$buildifier" -r testdir
"$buildifier" test.bzl
"$buildifier2" testdir/test.bzl > testdir/test.bzl.out

# directory without -r
"$buildifier" testdir > directory_error 2>&1 || true

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
cat > golden/directory_error <<EOF
buildifier: read testdir: is a directory
EOF

diff testdir/build golden/BUILD.golden
diff testdir/test.bzl golden/test.bzl.golden
diff testdir/subdir/test.bzl golden/test.bzl.golden
diff testdir/foo.bar golden/foo.bar
diff test.bzl golden/test.bzl.golden
diff stdout golden/test.bzl.golden
diff testdir/test.bzl.out golden/test.bzl.golden
diff directory_error golden/directory_error

# Test the linter

cat > testdir/to_fix.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar, cfg = "data")
EOF

cat > testdir/fixed_golden.bzl <<EOF
a = b // c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > testdir/fixed_golden_all.bzl <<EOF
a = b // c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > testdir/fixed_golden_dict_cfg.bzl <<EOF
a = b / c
d = {"a": 1, "b": 2}
attr.foo(bar)
EOF

cat > testdir/fixed_golden_cfg.bzl <<EOF
a = b / c
d = {"b": 2, "a": 1}
attr.foo(bar)
EOF

cat > testdir/fix_report_golden <<EOF
testdir/to_fix_tmp.bzl: applied fixes, 1 warnings left
fixed testdir/to_fix_tmp.bzl
EOF

error_docstring="testdir/to_fix_tmp.bzl:1: module-docstring: The file has no module docstring. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#module-docstring)"
error_integer="testdir/to_fix_tmp.bzl:1: integer-division: The \"/\" operator for integer division is deprecated in favor of \"//\". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division)"
error_dict="testdir/to_fix_tmp.bzl:2: unsorted-dict-items: Dictionary items are out of their lexicographical order. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#unsorted-dict-items)"
error_cfg="testdir/to_fix_tmp.bzl:3: attr-cfg: cfg = \"data\" for attr definitions has no effect and should be removed. (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#attr-cfg)"

test_lint () {
  ret=0
  cp testdir/to_fix.bzl testdir/to_fix_tmp.bzl
  echo "$4" > golden/error_golden

  cat > golden/fix_report_golden <<EOF
testdir/to_fix_tmp.bzl: applied fixes, $5 warnings left
fixed testdir/to_fix_tmp.bzl
EOF

  $buildifier --lint=warn $2 testdir/to_fix_tmp.bzl 2> testdir/error || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: warn: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff testdir/error golden/error_golden || die "$1: wrong console output for --lint=warn"

  $buildifier --lint=fix $2 -v testdir/to_fix_tmp.bzl 2> testdir/fix_report || ret=$?
  if [[ $ret -ne 4 ]]; then
    die "$1: fix: Expected buildifier to exit with 4, actual: $ret"
  fi
  diff testdir/to_fix_tmp.bzl $3 || die "$1: wrong file output for --lint=fix"
  diff testdir/fix_report golden/fix_report_golden || die "$1: wrong console output for --lint=fix"
}

test_lint "default" "" "testdir/fixed_golden.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_cfg" 1
test_lint "all" "--warnings=all" "testdir/fixed_golden_all.bzl" "$error_integer"$'\n'"$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
test_lint "cfg" "--warnings=attr-cfg" "testdir/fixed_golden_cfg.bzl" "$error_cfg" 0
test_lint "custom" "--warnings=-integer-division,+unsorted-dict-items" "testdir/fixed_golden_dict_cfg.bzl" "$error_docstring"$'\n'"$error_dict"$'\n'"$error_cfg" 1
