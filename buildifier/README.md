# Buildifier

buildifier is a tool for formatting bazel BUILD and .bzl files with a standard convention.

## Setup

Build the tool:
* Checkout the repo and then either via `go install` or `bazel build //buildifier`
* If you already have 'go' installed, then build a binary via:

`go get github.com/bazelbuild/buildtools/buildifier`

## Usage

Use buildifier to create standardized formatting for BUILD and .bzl files in the
same way that clang-format is used for source files.

    $ buildifier path/to/file

You can also process multiple files at once:

    $ buildifier path/to/file1 path/to/file2

You can make buildifier automatically find all Starlark files (i.e. BUILD, WORKSPACE, .bzl, or .sky)
in a directory recursively:

    $ buildifier -r path/to/dir

Buildifier supports the following file types: `BUILD`, `WORKSPACE`, `.bzl`, and
default, the latter is reserved for Starlark files buildifier doesn't know about
(e.g. configuration files for third-party projects that use Starlark). The
formatting rules for WORKSPACE files are the same as for BUILD files (both are
declarative and have stricter formatting rules), and default files are formatted
similarly to .bzl files, allowing more flexibility. Different linter warnings
may be limited to any subset of these file types, e.g. a certain warning may be
only relevant to Bazel files (i.e. `BUILD`, `WORKSPACE`, and `.bzl`) or to
non-WORKSPACE files.

Buildifier automatically detects the file type by its filename, taking into
account optional prefixes and suffixes, e.g. `BUILD`, `BUILD.oss`, or
`BUILD.bazel` will be detected as BUILD files, and `build_defs.bzl.oss` is a
.bzl file. Files with unknown names (e.g. `foo.bar`) or files passed via stdin
will be treated as default file type. To override the automatic file type
detection use the `--type` flag explicitly:

    $ cat foo.bar | buildifier --type=build
    $ cat foo.bar | buildifier --type=bzl
    $ cat foo.bar | buildifier --type=workspace
    $ cat foo.bar | buildifier --type=default

## Linter

Buildifier has an integrated linter that can point out and in some cases
automatically fix various issues. To use it launch one of the following commands
to show and to fix the issues correspondingly (note that some issues cannot be
fixed automatically):

    buildifier --lint=warn path/to/file
    buildifier --lint=fix path/to/file

By default, the linter searches for all known issues relevant for the given
file type except those that are marked with
"[Disabled by default](../WARNINGS.md)" in the documentation.

You can specify the categories using the `--warnings` flag either by providing
the categories explicitly:

    buildifier --lint=warn --warnings=positional-args,duplicated-name

or by modifying the default warnings set by using `+` or `-` modifiers before
each warning category:

    buildifier --lint=warn --warnings=-positional-args,+unsorted-dict-items

It's also possible to provide `--warnings=all` to use all supported warnings
categories (they will still be limited to relevant warnings for the given file
type).

See also the [full list](../WARNINGS.md) or the supported warnings.

## Setup and usage via Bazel (not supported on Windows)

You can also invoke buildifier via the Bazel rule.
`WORKSPACE` file:
```bzl
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# buildifier is written in Go and hence needs rules_go to be built.
# See https://github.com/bazelbuild/rules_go for the up to date setup instructions.
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "d1ffd055969c8f8d431e2d439813e42326961d0942bdf734d2c95dc30c369566",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.24.5/rules_go-v0.24.5.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.24.5/rules_go-v0.24.5.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    sha256 = "b85f48fa105c4403326e9525ad2b2cc437babaa6e15a3fc0b1dbab0ab064bc7c",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.2/bazel-gazelle-v0.22.2.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.2/bazel-gazelle-v0.22.2.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

# If you use WORKSPACE.bazel, use the following line instead of the bare gazelle_dependencies():
# gazelle_dependencies(go_repository_default_config = "@//:WORKSPACE.bazel")
gazelle_dependencies()

http_archive(
    name = "com_google_protobuf",
    strip_prefix = "protobuf-master",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/master.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    strip_prefix = "buildtools-master",
    url = "https://github.com/bazelbuild/buildtools/archive/master.zip",
)
```

`BUILD.bazel` typically in the workspace root:
```bzl
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(
    name = "buildifier",
)
```
Invoke with
```bash
bazel run //:buildifier
```

## File diagnostics in json

Buildifier supports diagnostics output in machine-readable format (json), triggered by
`--format=json` (only works in combination with `--mode=check`). If used in combination with `-v`,
the output json will be indented for better readability.

The output format is the following:

```json
{
    "success": false,  // true if all files are formatted and generate no warnings, false otherwise
    "files": [  // list of all files processed by buildifier
        {
            "filename": "file_1.bzl",
            "formatted": true,  // whether the file is correctly formatted
            "valid": true,  // whether the file is a valid Starlark file. Can only be false if formatted = false
            "warnings": [  // a list of warnings
                {
                    "start": {
                        "line": 1,
                        "column": 5
                    },
                    "end": {
                        "line": 1,
                        "column": 10
                    },
                    "category": "integer-division",
                    "actionable": true,
                    "message": "The \"/\" operator for integer division is deprecated in favor of \"//\".",
                    "url": "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#integer-division"
                }
            ]
        },
        {
            "filename": "file_2.bzl",
            "formatted": false,
            "valid": true,
            "warnings": [],
            "rewrites": {  // technical information, a list of rewrites buildifier applies during reformatting
                "editoctal": 1
            }
        },
        {
            "filename": "file_3.bzl",
            "formatted": true,
            "valid": true,
            "warnings": []
        },
        {
            "filename": "file_4.not_bzl",
            "formatted": false,
            "valid": false,
            "warnings": []
        }
    ]
}
```

When the `--format` flag is provided, buildifier always returns `0` unless there are internal
failures or wrong input parameters, this means the output can be parsed as JSON, and its `success`
field should be used to determine whether the diagnostics result is positive.
