set -e

INPUT="def f(): pass"  # formatted differently in build and bzl modes
BUILD_OUTPUT="def f(): pass"
BZL_OUTPUT="def f():\n    pass"

mkdir test
echo -e "$INPUT" > test/BUILD
echo -e "$INPUT" > test/test.bzl

$2 test/test.bzl > test/test.bzl.out

$1 --type=auto test/*

echo -e "$BUILD_OUTPUT" > test/BUILD.golden
echo -e "$BZL_OUTPUT" > test/test.bzl.golden

diff test/BUILD test/BUILD.golden
diff test/test.bzl test/test.bzl.golden

diff test/test.bzl.out test/test.bzl.golden
