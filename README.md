# Buildifier
buildifier is a tool for formatting bazel BUILD files with a standard convention.

## Setup

Build the tool:
* Checkout the repo and then either via `go install` or `bazel build //buildifier`
* If you already have 'go' installed, then build a binary via: 

`go install github.com/bazelbuild/buildifier/buildifier`

## Usage

Use buildifier to create standardized formatting for BUILD files in the
same way that clang-format is used for source files.

`$ buildifier -showlog -mode=check $(find . -iname BUILD -type f)`
