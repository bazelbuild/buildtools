#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/test_utils.sh"

init_buildifier "$1"
setup_test_env

# Tests handling of multiple files and dependencies (e.g. deprecated functions).

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
BUILD:1: deprecated-function: The function "bar" defined in "//lib.bzl" is deprecated. (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#deprecated-function)
EOF

$buildifier --lint=warn --warnings=deprecated-function BUILD 2> report || ret=$?
diff -u report_golden report || die "$1: wrong console output for multifile warnings (WORKSPACE exists)"

cd ..
