# buildifier script for formatting bazel BUILD files

## Setup

Build the tool:
* Checkout the repo and then either via `go install` or `bazel build //buildifier`
* If you already have 'go' installed, then build a binary via: 

`go get -d -u github.com/bazelbuild/buildifier/buildifier && go generate github.com/bazelbuild/buildifier/core && go install github.com/bazelbuild/buildifier/buildifier`

Note: the extra 'generate' step as 'go get' does not run 'generate' when installing.

## Usage

Use buildifier to create standardized formatting for BUILD files in the
same way that clang-format is used for source files.
