#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/test_utils.sh"

init_buildifier "$1"
setup_test_env

# Tests allowed symbol load locations check.

mkdir test_dir/allowed_locations
cd test_dir/allowed_locations

cat > BUILD <<EOF
load(":f.bzl", "s1", "s2")
load(":a.bzl", "s3")
load(":a.bzl", "s4")
EOF

cat > buildifier.tables <<EOF
{
  "AllowedSymbolLoadLocations": {
    "s1": [":z.bzl"],
    "s3": [":y.bzl", ":x.bzl"],
    "s4": [":a.bzl"]
  }
}
EOF

cat > report_golden <<EOF
BUILD:1: allowed-symbol-load-locations: Symbol "s1" must be loaded from :z.bzl. (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#allowed-symbol-load-locations)
BUILD:2: allowed-symbol-load-locations: Symbol "s3" must be loaded from one of the allowed locations: :x.bzl, :y.bzl. (https://github.com/bazelbuild/buildtools/blob/main/WARNINGS.md#allowed-symbol-load-locations)
EOF

$buildifier --lint=warn --warnings=allowed-symbol-load-locations -tables=buildifier.tables BUILD 2> report || true
diff -u report_golden report || die "$1: wrong console output for allowed symbol load locations"

cd ../..
