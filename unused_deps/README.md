# Unused Deps

unused_deps is a command line tool to determine any unused depenencies
in java_library targets in [Bazel](https://github.com/bazelbuild/bazel)
BUILD files using.  It outputs `buildozer` commands to apply the suggested
pruning.

## Dependencies

1. Protobuf go runtime: to download (if not using bazel)
`go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`


## Installation

1. Change directory to the buildifier/unused_deps

```bash
gopath=$(go env GOPATH)
cd $gopath/src/github.com/bazelbuild/buildtools/unused_deps
```

2. Install

```bash
go install
```

## Usage

```shell
unused_deps TARGET...
```

Here, `TARGET` is a space-separated list of Bazel labels, with support for ...
