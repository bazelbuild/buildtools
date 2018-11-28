set -e

mkdir test
INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > test/build  # case doesn't matter
echo -e "$INPUT" > test/test.bzl

$1 < test/build > stdout
$1 test/*
$2 test/test.bzl > test/test.bzl.out

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
EOF

cat > test/error_golden <<EOF
test/to_fix.bzl:1: integer-division: The "/" operator for integer division is deprecated in favor of "//". (https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division)
EOF

cat > test/fixed_golden.bzl <<EOF
a = b // c
EOF

cat > test/fix_report_golden <<EOF
test/to_fix.bzl: applied fixes, 0 warnings left
fixed test/to_fix.bzl
EOF

ret=0
$1 --lint=warn test/to_fix.bzl 2> test/error || ret=$?
if [[ $ret -ne 4 ]]; then
  echo "Expected buildifier to exit with 4" >&2
  exit 1
fi
diff test/error test/error_golden

$1 --lint=fix -v test/to_fix.bzl 2> test/fix_report || ret=$?
if [[ $ret -ne 4 ]]; then
  echo "Expected buildifier to exit with 4" >&2
  exit 1
fi
diff test/to_fix.bzl test/fixed_golden.bzl
diff test/fix_report test/fix_report_golden
