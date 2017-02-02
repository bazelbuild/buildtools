# Bazel tools

This repo contains two useful tools for working with bazel BUILD files

* buildifier is a tool for formatting bazel BUILD files with a standard convention.
* buildozer is a tool for programmatic or command-line editing of BUILD files.


## Setup

Binary downloads for linux/Mac are available on the releases page.
(Any help with Windows appreciated.)

Build the tool from source:
* Checkout the repo and then 
`bazel build //buildifier //buildozer`
* If you already have 'go' installed, then build a binary via: 

`go install github.com/bazelbuild/buildifier/buildifier`

`go install github.com/bazelbuild/buildifier/buildozer`

## Buildifier Usage

Use buildifier to create standardized formatting for BUILD files in the
same way that clang-format is used for source files.

`$ buildifier -showlog -mode=check $(find . -iname BUILD -type f)`

## Buildozer Usage

e.g.
`buildozer 'add deps @jsr305_maven//jar' //java/org/my/project:lib`

Will add jsr305_maven in the deps attribute of the "lib" rule in the BUILD file
at that location.
