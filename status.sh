#!/bin/bash

set -e

buildifier_tags=$(git describe --tags)
IFS='-' read -a parse_tags <<< "$buildifier_tags"
echo "BUILDIFIER_VERSION ${parse_tags[0]}"

buildifier_rev=$(git rev-parse HEAD)
echo "BUILD_SCM_REVISION ${buildifier_rev}"
