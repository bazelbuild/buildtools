# Buildtools for bazel

This repository contains developer tools for working with Google's `bazel` buildtool.

* [buildifier](buildifier/README.md) For formatting BUILD, BUILD.bazel and BUCK files in a standard way
* [buildozer](buildozer/README.md) For doing command-line operations on these files.
* [unused_deps](unused_deps/README.md) For finding unneeded dependencies in
[java_library](https://docs.bazel.build/versions/master/be/java.html#java_library) rules.


linux-x86_64 | ubuntu_15.10-x86_64 | darwin-x86_64
:---: | :---: | :---:
[![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=linux-x86_64)](http://ci.bazel.io/job/buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=linux-x86_64) | [![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=ubuntu_15.10-x86_64)](http://ci.bazel.io/job/buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=ubuntu_15.10-x86_64) | [![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=darwin-x86_64)](http://ci.bazel.io/job/buildtools/BAZEL_VERSION=latest,PLATFORM_NAME=darwin-x86_64)

## Setup

See instructions in each tool's directory.
