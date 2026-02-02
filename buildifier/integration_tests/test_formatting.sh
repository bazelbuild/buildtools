#!/bin/bash

set -euo pipefail

source "$(dirname "$0")/test_utils.sh"

init_buildifier "$1"
setup_test_env

# Tests default formatting behavior, recursive formatting, and configuration file usage.

INPUT="load(':foo.bzl', 'foo'); foo(tags=['b', 'a'],srcs=['d', 'c'])"  # formatted differently in build and bzl modes
echo -e "$INPUT" > test_dir/BUILD
echo -e "$INPUT" > test_dir/test.bzl
echo -e "$INPUT" > test_dir/subdir/test.bzl
echo -e "$INPUT" > test_dir/subdir/build  # lowercase, should be ignored by -r
echo -e "$INPUT" > test.bzl  # outside the test_dir directory
echo -e "$INPUT" > test2.bzl  # outside the test_dir directory
echo -e "not valid +" > test_dir/foo.bar
echo -e '{ "type": "build" }' > test_dir/.buildifier.test.json # demonstrate config file works by overriding input type to format a bzl file as a BUILD file.
mkdir test_dir/workspace  # name of a starlark file, but a directory
mkdir test_dir/.git  # contents should be ignored
echo -e "a+b" > test_dir/.git/git.bzl
cat > test_dir/MODULE.bazel <<'EOF'
module(name='my-module',version='1.0',compatibility_level=1)
include("cpp.MODULE.bazel")

include("go.MODULE.bazel")
include("web.MODULE.bazel")
bazel_dep(name='rules_cc',version='0.0.1')
bazel_dep(name='protobuf',repo_name='com_google_protobuf',version='3.19.0')
bazel_dep(
    name='rules_go',
    version='0.37.0',
    repo_name='io_bazel_rules_go',
)
go_sdk=use_extension("@io_bazel_rules_go//go:extensions.bzl","go_sdk")
# Known to exist since it is instantiated by rules_go itself.
use_repo(go_sdk,"go_default_sdk")
non_module_deps = use_extension("//internal/bzlmod:non_module_deps.bzl","non_module_deps")
use_repo(
    non_module_deps,
    "bazel_gazelle_go_repository_tools",
    "bazel_gazelle_go_repository_config",
    "bazel_gazelle_go_repository_cache",
)
rules_go_non_module_deps = use_extension("@io_bazel_rules_go//go/private:extensions.bzl","non_module_dependencies",dev_dependency=True)
use_repo(rules_go_non_module_deps,"go_googleapis")
go_deps  =  use_extension("//:extensions.bzl",  "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
use_repo(
    go_deps,
    "com_github_fsnotify_fsnotify",
    "com_github_fsnotify_fsnotify",
    "com_github_bmatcuk_doublestar_v4",
    "buildtools",
    "com_github_google_go_cmp",
    "com_github_pelletier_go_toml",
    "org_golang_x_mod",
    "com_github_pmezard_go_difflib",
    # Separated by comment.
    "org_golang_x_sync",
    "org_golang_x_tools",
    # Used internally by the go_deps module extension.
    "bazel_gazelle_go_repository_directives",
    c = "a",
    b = "b",
    a = "c",
)
go_deps.module(name = "foo")
use_repo(go_deps, "foo")
bazel_dep(name="foo",version="1.0")
git_override(module_name="foo",remote="foo.git",commit="1234567890")
bazel_dep(name="bar",version="1.0")
archive_override(module_name="not_bar",integrity="sha256-1234567890")
# do not sort
use_repo(go_deps, "b", "b", "a")
use_repo(
    # do not sort
    go_deps,
    "b",
    "b",
    "a",
)
use_repo(
    go_deps,
    # do not sort
    "b",
    "b",
    "a",
)

bazel_dep(name='prod_dep',version='3.19.0')
bazel_dep(name='other_prod_dep',version='3.19.0',dev_dependency=False)
bazel_dep(name='dev_dep',version='3.19.0',dev_dependency=True)
bazel_dep(name = "weird_dep", version = "3.19.0", dev_dependency = "True" == "True")
bazel_dep(name='yet_another_prod_dep',version='3.19.0')
EOF

cp test_dir/foo.bar golden/foo.bar
cp test_dir/subdir/build golden/build
cp test_dir/.git/git.bzl golden/git.bzl

"$buildifier" < test_dir/BUILD > stdout
"$buildifier" -r test_dir
"$buildifier" test.bzl
"$buildifier" --path=foo.bzl test2.bzl
"$buildifier" --config=test_dir/.buildifier.test.json < test_dir/test.bzl > test_dir/test.bzl.BUILD.out
"$buildifier" --config=example > test_dir/.buildifier.example.json

cat > golden/BUILD.golden <<EOF
load(":foo.bzl", "foo")

foo(
    srcs = [
        "c",
        "d",
    ],
    tags = [
        "a",
        "b",
    ],
)
EOF

cat > golden/test.bzl.golden <<EOF
load(":foo.bzl", "foo")

foo(tags = ["b", "a"], srcs = ["d", "c"])
EOF

cat > golden/MODULE.bazel.golden <<EOF
module(
    name = "my-module",
    version = "1.0",
    compatibility_level = 1,
)

include("cpp.MODULE.bazel")
include("go.MODULE.bazel")
include("web.MODULE.bazel")

bazel_dep(name = "rules_cc", version = "0.0.1")
bazel_dep(name = "protobuf", version = "3.19.0", repo_name = "com_google_protobuf")
bazel_dep(
    name = "rules_go",
    version = "0.37.0",
    repo_name = "io_bazel_rules_go",
)

go_sdk = use_extension("@io_bazel_rules_go//go:extensions.bzl", "go_sdk")

# Known to exist since it is instantiated by rules_go itself.
use_repo(go_sdk, "go_default_sdk")

non_module_deps = use_extension("//internal/bzlmod:non_module_deps.bzl", "non_module_deps")
use_repo(
    non_module_deps,
    "bazel_gazelle_go_repository_cache",
    "bazel_gazelle_go_repository_config",
    "bazel_gazelle_go_repository_tools",
)

rules_go_non_module_deps = use_extension("@io_bazel_rules_go//go/private:extensions.bzl", "non_module_dependencies", dev_dependency = True)
use_repo(rules_go_non_module_deps, "go_googleapis")

go_deps = use_extension("//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
use_repo(
    go_deps,
    "buildtools",
    "com_github_bmatcuk_doublestar_v4",
    "com_github_fsnotify_fsnotify",
    "com_github_google_go_cmp",
    "com_github_pelletier_go_toml",
    "com_github_pmezard_go_difflib",
    "org_golang_x_mod",
    # Separated by comment.
    "org_golang_x_sync",
    "org_golang_x_tools",
    # Used internally by the go_deps module extension.
    "bazel_gazelle_go_repository_directives",
    a = "c",
    b = "b",
    c = "a",
)

go_deps.module(name = "foo")
use_repo(go_deps, "foo")

bazel_dep(name = "foo", version = "1.0")
git_override(
    module_name = "foo",
    commit = "1234567890",
    remote = "foo.git",
)

bazel_dep(name = "bar", version = "1.0")

archive_override(
    module_name = "not_bar",
    integrity = "sha256-1234567890",
)

# do not sort
use_repo(go_deps, "b", "a")
use_repo(
    # do not sort
    go_deps,
    "b",
    "a",
)
use_repo(
    go_deps,
    # do not sort
    "b",
    "a",
)

bazel_dep(name = "prod_dep", version = "3.19.0")
bazel_dep(name = "other_prod_dep", version = "3.19.0", dev_dependency = False)

bazel_dep(name = "dev_dep", version = "3.19.0", dev_dependency = True)
bazel_dep(name = "weird_dep", version = "3.19.0", dev_dependency = "True" == "True")

bazel_dep(name = "yet_another_prod_dep", version = "3.19.0")
EOF

cat > golden/.buildifier.example.json <<EOF
{
  "type": "auto",
  "mode": "fix",
  "lint": "fix",
  "warningsList": [
    "allowed-symbol-load-locations",
    "attr-applicable_licenses",
    "attr-cfg",
    "attr-license",
    "attr-licenses",
    "attr-non-empty",
    "attr-output-default",
    "attr-single-file",
    "build-args-kwargs",
    "bzl-visibility",
    "canonical-repository",
    "confusing-name",
    "constant-glob",
    "ctx-actions",
    "ctx-args",
    "deprecated-function",
    "depset-items",
    "depset-iteration",
    "depset-union",
    "dict-concatenation",
    "dict-method-named-arg",
    "duplicated-name",
    "external-path",
    "filetype",
    "function-docstring",
    "function-docstring-args",
    "function-docstring-header",
    "function-docstring-return",
    "git-repository",
    "http-archive",
    "integer-division",
    "keyword-positional-params",
    "list-append",
    "load",
    "module-docstring",
    "name-conventions",
    "native-android",
    "native-build",
    "native-cc-binary",
    "native-cc-common",
    "native-cc-debug-package-info",
    "native-cc-fdo-prefetch-hints",
    "native-cc-fdo-profile",
    "native-cc-import",
    "native-cc-info",
    "native-cc-library",
    "native-cc-memprof-profile",
    "native-cc-objc-import",
    "native-cc-objc-library",
    "native-cc-propeller-optimize",
    "native-cc-proto",
    "native-cc-shared-library",
    "native-cc-shared-library-hint-info",
    "native-cc-shared-library-info",
    "native-cc-test",
    "native-cc-toolchain",
    "native-cc-toolchain-suite",
    "native-java-binary",
    "native-java-common",
    "native-java-import",
    "native-java-info",
    "native-java-library",
    "native-java-lite-proto",
    "native-java-package-config",
    "native-java-plugin",
    "native-java-plugin-info",
    "native-java-proto",
    "native-java-runtime",
    "native-java-test",
    "native-java-toolchain",
    "native-package",
    "native-proto",
    "native-proto-common",
    "native-proto-info",
    "native-proto-lang-toolchain",
    "native-proto-lang-toolchain-info",
    "native-py",
    "native-sh-binary",
    "native-sh-library",
    "native-sh-test",
    "no-effect",
    "output-group",
    "overly-nested-depset",
    "package-name",
    "package-on-top",
    "positional-args",
    "print",
    "provider-params",
    "redefined-variable",
    "repository-name",
    "return-value",
    "rule-impl-return",
    "skylark-comment",
    "skylark-docstring",
    "string-iteration",
    "uninitialized",
    "unnamed-macro",
    "unreachable",
    "unsorted-dict-items",
    "unused-variable"
  ]
}
EOF

diff -u test_dir/BUILD golden/BUILD.golden
diff -u test_dir/test.bzl golden/test.bzl.golden
diff -u test_dir/subdir/test.bzl golden/test.bzl.golden
diff -u test_dir/test.bzl.BUILD.out golden/BUILD.golden
diff -u test_dir/subdir/build golden/build
diff -u test_dir/foo.bar golden/foo.bar
diff -u test.bzl golden/test.bzl.golden
diff -u test2.bzl golden/test.bzl.golden
diff -u stdout golden/test.bzl.golden
diff -u test_dir/.git/git.bzl golden/git.bzl
diff -u test_dir/MODULE.bazel golden/MODULE.bazel.golden
diff -u test_dir/.buildifier.example.json golden/.buildifier.example.json

# Test run on a directory without -r
"$buildifier" test_dir || ret=$?
if [[ $ret -ne 3 ]]; then
  die "Directory without -r: expected buildifier to exit with 3, actual: $ret"
fi
