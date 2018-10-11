#!/usr/bin/env bash

set -eux

for r in $(bazel query 'kind(sh_binary, //...)' 2> /dev/null | grep _copy | xargs)
do
  bazel run $r
done
