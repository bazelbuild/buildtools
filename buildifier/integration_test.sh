set -e

INPUT="foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
BUILD_OUTPUT="foo(\n    srcs = [\n        \"c\",\n        \"d\",\n    ],\n    tags = [\n        \"a\",\n        \"b\",\n    ],\n)"
BZL_OUTPUT="foo(\n    tags = [\n        \"b\",\n        \"a\",\n    ],\n    srcs = [\n        \"d\",\n        \"c\",\n    ],\n)"

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
