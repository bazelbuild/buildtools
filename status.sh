#!/bin/bash

buildifier_tags=$(git describe --tags)
IFS='-' read -a parse_tags <<< "$buildifier_tags"
if [[ $? != 0 ]];
then
    exit 1
fi
echo "BUILDIFIER_VERSION ${parse_tags[0]}"


buildifier_rev=$(git rev-parse HEAD)
if [[ $? != 0 ]];
then
    exit 1
fi

echo "BUILD_SCM_REVISION ${buildifier_rev}"
