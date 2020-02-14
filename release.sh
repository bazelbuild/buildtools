#!/usr/bin/env bash

set -u -e -o pipefail 

# TODO(vladmos): add bits for publishing go binaries to github￼

# Googlers: you should npm login using the go/npm-publish service:
#      $ npm login --registry https://wombat-dressing-room.appspot.com
# Non-googlers: you should run this script with
#      $ NPM_REGISTRY=https://registry.npmjs.org ./release.sh ...
REGISTRY=${NPM_REGISTRY:-https://wombat-dressing-room.appspot.com}
readonly NPM_ARGS=(
    "--access public"
    "--tag latest"
    "--registry $REGISTRY"
    # Uncomment for testing
    # "--dry-run"
)
for pkg in buildifier buildozer; do
    bazel run --config=release //$pkg:npm_package.publish -- ${NPM_ARGS[@]}
done