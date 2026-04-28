#!/usr/bin/env bash

set -e

if [ -z "$GITHUB_ACCESS_TOKEN" ]
then
    echo "Create a Github auth token at https://github.com/settings/tokens and set it as \$GITHUB_ACCESS_TOKEN"
    exit 1
fi

TAG="$(git describe --abbrev=0 --tags | sed 's/* //')"  # v1.2.3
VERSION=$(echo $TAG | sed 's/v//')  # 1.2.3
DATE="$(date +"%Y-%m-%d")"
NAME="Release $VERSION ($DATE)"

GH_REPO="repos/bazelbuild/buildtools"
GH_AUTH_HEADER="Authorization: token $GITHUB_ACCESS_TOKEN"

BIN_DIR=`mktemp -d -p "$DIR"`

bazel clean
bazel build --config=release //buildifier:all //buildozer:all //unused_deps:all
  
#######################################
# Copies all present output binaries to the given directory.
# Arguments:
#   Directory path to copy binaries to
# Returns:
#   0 if successful, non-zero on error
#######################################
copy_go_binaries() {
  local out_dir="$1"
  # Lists all GoLink actions produced by go_binary rules in the repository.
  local go_binary_outputs="$(bazel aquery --config=release \
      'mnemonic("GoLink", kind(go_binary, ...))' \
      --noinclude_aspects \
      --noinclude_commandline)"

  while IFS= read -r line ; do 
    # Parses out the "Outputs" and copies them to the output dir.
    local binary_output_path=$(echo "$line" | sed -n 's/^.*Outputs: \[\(.*\)\]/\1/p')
    if [[ ! -z "$binary_output_path" ]]; then
      # Ignores errors from "cp" since aquery will include some binaries besides
      # the expected tools. This script later validates that all required binaries
      # are present.
      cp "$binary_output_path" "$out_dir" 2>/dev/null || true
    fi
  done <<< "$go_binary_outputs"
}
  
echo "Copies binaries to temp dir"
copy_go_binaries "$BIN_DIR"

# The list of tools to include in release.
tools=("buildifier" "buildozer" "unused_deps")
# Binary suffixes which should be included for all tools.
binary_target_suffixes=(
  "linux_amd64"
  "linux_arm64"
  "linux_riscv64"
  "linux_s390x"
  "darwin_amd64"
  "darwin_arm64"
  "windows_amd64.exe"
  "windows_arm64.exe"
)
# Generates list of all $tool-$suffix binaries which should be included.
all_binary_names=()
for tool in "${tools[@]}"; do
  for binary_suffix in "${binary_target_suffixes[@]}"; do
    all_binary_names+=("$tool-$binary_suffix")
  done
done

echo "Validating that all expected binaries are present"
for binary_name in "${all_binary_names[@]}"; do
  if [[ ! -f "$BIN_DIR/$binary_name" ]]; then
    echo "Expected binary \"$binary_name\" was not found"
    exit 2
  fi
done

echo "Creating a draft release"
API_JSON="{\"tag_name\": \"$TAG\", \"target_commitish\": \"main\", \"name\": \"$NAME\", \"draft\": true}"
RESPONSE=$(curl -s --show-error -H "$GH_AUTH_HEADER" --data "$API_JSON" "https://api.github.com/$GH_REPO/releases")
RELEASE_ID=$(echo $RESPONSE | jq -r '.id')
RELEASE_URL=$(echo $RESPONSE | jq -r '.html_url')

upload_file() {
    echo "Uploading $2"
    ASSET="https://uploads.github.com/$GH_REPO/releases/$RELEASE_ID/assets?name=$2"
    curl --data-binary @"$1" -s --show-error -o /dev/null -H "$GH_AUTH_HEADER" -H "Content-Type: application/octet-stream" $ASSET
}

for binary_name in "${all_binary_names[@]}"; do
  # Output should contain dashes instead of underscores.
  output_name=${binary_name//_/-}
  upload_file "$BIN_DIR/$binary_name" "$output_name"
done

rm -rf $BIN_DIR

echo "The draft release is available at $RELEASE_URL"
