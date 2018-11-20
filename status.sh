#!/usr/bin/env bash

set -e

buildifier_tags=$(git describe --tags)
IFS='-' read -a parse_tags <<< "$buildifier_tags"
echo "STABLE_buildVersion ${parse_tags[0]}"

buildifier_rev=$(git rev-parse HEAD)
echo "STABLE_buildScmRevision ${buildifier_rev}"
