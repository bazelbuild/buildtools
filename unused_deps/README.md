# Unused Deps

unused_deps is a command line tool to determine any unused dependencies
in [java_library](https://docs.bazel.build/versions/master/be/java.html#java_library)
rules. targets.  It outputs `buildozer` commands to apply the suggested
prunings.

## Installation

Build a binary and put it into your $GOPATH/bin:

```bash
go get github.com/bazelbuild/buildtools/unused_deps
```

## Usage

```shell
unused_deps TARGET...
```

Here, `TARGET` is a space-separated list of Bazel labels, with support for `:all` and `...`
