/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package warn

import (
	"fmt"
	"testing"

	"github.com/bazelbuild/buildtools/tables"
)

func TestAttrDataConfigurationWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-cfg", `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, cfg = "data"),
  }
)

attr.label_list(mandatory = True, cfg = "exec")`, `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True),
  }
)

attr.label_list(mandatory = True, cfg = "exec")`,
		[]string{`:3: cfg = "data" for attr definitions has no effect and should be removed.`},
		scopeBzl)
}

func TestAttrHostConfigurationWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-cfg", `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, cfg = "host"),
  }
)

attr.label_list(mandatory = True, cfg = "exec")`, `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, cfg = "exec"),
  }
)

attr.label_list(mandatory = True, cfg = "exec")`,
		[]string{`:3: cfg = "host" for attr definitions should be replaced by cfg = "exec".`},
		scopeBzl)
}

func TestDepsetItemsWarning(t *testing.T) {
	checkFindings(t, "depset-items", `
def f():
  depset(items=foo)
  a = depset()
  depset(a)
`, []string{
		`:2: Parameter "items" is deprecated, use "direct" and/or "transitive" instead.`,
		`:4: Giving a depset as first unnamed parameter to depset() is deprecated, use the "transitive" parameter instead.`,
	},
		scopeEverywhere)
}

func TestAttrNonEmptyWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-non-empty", `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, non_empty = True),
      "bar": attr.label_list(mandatory = True, non_empty = False),
      "baz": attr.label_list(mandatory = True, non_empty = foo.bar()),
      "qux": attr.label_list(mandatory = True, non_empty = not foo.bar()),
      "aaa": attr.label_list(mandatory = True, non_empty = (foo.bar())),
      "bbb": attr.label_list(mandatory = True, non_empty = (
					not foo.bar()
			)),
  }
)`, `
rule(
  attrs = {
      "foo": attr.label_list(mandatory = True, allow_empty = False),
      "bar": attr.label_list(mandatory = True, allow_empty = True),
      "baz": attr.label_list(mandatory = True, allow_empty = not foo.bar()),
      "qux": attr.label_list(mandatory = True, allow_empty = foo.bar()),
      "aaa": attr.label_list(mandatory = True, allow_empty = (not foo.bar())),
      "bbb": attr.label_list(mandatory = True, allow_empty = (
					foo.bar()
			)),
  }
)`,
		[]string{
			":3: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":4: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":5: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":6: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":7: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
			":8: non_empty attributes for attr definitions are deprecated in favor of allow_empty.",
		},
		scopeBzl)
}

func TestAttrSingleFileWarning(t *testing.T) {
	checkFindingsAndFix(t, "attr-single-file", `
rule(
  attrs = {
      "foo": attr.label_list(single_file = True, allow_files = [".cc"], mandatory = True),
      "bar": attr.label_list(single_file = True, mandatory = True),
      "baz": attr.label_list(single_file = False, mandatory = True),
  }
)`, `
rule(
  attrs = {
      "foo": attr.label_list(allow_single_file = [".cc"], mandatory = True),
      "bar": attr.label_list(allow_single_file = True, mandatory = True),
      "baz": attr.label_list(mandatory = True),
	}
)`,
		[]string{
			":3: single_file is deprecated in favor of allow_single_file.",
			":4: single_file is deprecated in favor of allow_single_file.",
			":5: single_file is deprecated in favor of allow_single_file.",
		},
		scopeBzl)
}

func TestCtxActionsWarning(t *testing.T) {
	checkFindingsAndFix(t, "ctx-actions", `
def impl(ctx):
  ctx.new_file(foo)
  ctx.new_file(foo, "foo %s " % bar)
  ctx.new_file(foo, name = bar)
  ctx.new_file(foo, bar, baz)
  ctx.experimental_new_directory(foo, bar)
  ctx.file_action(foo, bar)
  ctx.file_action(foo, bar, executable = True)
  ctx.action(foo, bar, command = "foo")
  ctx.action(foo, bar, executable = "bar")
  ctx.empty_action(foo, bar)
  ctx.template_action(foo, bar)
  ctx.template_action(foo, bar, executable = True)
	ctx.foobar(foo, bar)
`, `
def impl(ctx):
  ctx.actions.declare_file(foo)
  ctx.actions.declare_file("foo %s " % bar, sibling = foo)
  ctx.actions.declare_file(bar, sibling = foo)
  ctx.new_file(foo, bar, baz)
  ctx.actions.declare_directory(foo, bar)
  ctx.actions.write(foo, bar)
  ctx.actions.write(foo, bar, is_executable = True)
  ctx.actions.run_shell(foo, bar, command = "foo")
  ctx.actions.run(foo, bar, executable = "bar")
  ctx.actions.do_nothing(foo, bar)
  ctx.actions.expand_template(foo, bar)
  ctx.actions.expand_template(foo, bar, is_executable = True)
	ctx.foobar(foo, bar)
`,
		[]string{
			`:2: "ctx.new_file" is deprecated in favor of "ctx.actions.declare_file".`,
			`:3: "ctx.new_file" is deprecated in favor of "ctx.actions.declare_file".`,
			`:4: "ctx.new_file" is deprecated in favor of "ctx.actions.declare_file".`,
			`:5: "ctx.new_file" is deprecated in favor of "ctx.actions.declare_file".`,
			`:6: "ctx.experimental_new_directory" is deprecated in favor of "ctx.actions.declare_directory".`,
			`:7: "ctx.file_action" is deprecated in favor of "ctx.actions.write".`,
			`:8: "ctx.file_action" is deprecated in favor of "ctx.actions.write".`,
			`:9: "ctx.action" is deprecated in favor of "ctx.actions.run_shell".`,
			`:10: "ctx.action" is deprecated in favor of "ctx.actions.run".`,
			`:11: "ctx.empty_action" is deprecated in favor of "ctx.actions.do_nothing".`,
			`:12: "ctx.template_action" is deprecated in favor of "ctx.actions.expand_template".`,
			`:13: "ctx.template_action" is deprecated in favor of "ctx.actions.expand_template".`,
		},
		scopeBzl)

	checkFindings(t, "ctx-actions", `
def impl(ctx):
  ctx.new_file(foo, bar, baz)
`, []string{
		`:2: "ctx.new_file" is deprecated in favor of "ctx.actions.declare_file".`,
	}, scopeBzl)
}

func TestPackageNameWarning(t *testing.T) {
	checkFindingsAndFix(t, "package-name", `
foo(a = PACKAGE_NAME)

def f(PACKAGE_NAME):
    foo(a = PACKAGE_NAME)

def g():
    foo(a = PACKAGE_NAME)
`, `
foo(a = native.package_name())

def f(PACKAGE_NAME):
    foo(a = PACKAGE_NAME)

def g():
    foo(a = native.package_name())
`,
		[]string{
			`:1: Global variable "PACKAGE_NAME" is deprecated in favor of "native.package_name()". Please rename it.`,
			`:7: Global variable "PACKAGE_NAME" is deprecated in favor of "native.package_name()". Please rename it.`,
		},
		scopeBzl)

	checkFindings(t, "package-name", `
PACKAGE_NAME = "foo"
foo(a = PACKAGE_NAME)
`, []string{}, scopeBzl)

	checkFindings(t, "package-name", `
load(":foo.bzl", "PACKAGE_NAME")
foo(a = PACKAGE_NAME)
`, []string{}, scopeBzl)
}

func TestRepositoryNameWarning(t *testing.T) {
	checkFindingsAndFix(t, "repository-name", `
foo(a = REPOSITORY_NAME)

def f(REPOSITORY_NAME):
    foo(a = REPOSITORY_NAME)

def g():
    foo(a = REPOSITORY_NAME)
`, `
foo(a = native.repository_name())

def f(REPOSITORY_NAME):
    foo(a = REPOSITORY_NAME)

def g():
    foo(a = native.repository_name())
`,
		[]string{
			`:1: Global variable "REPOSITORY_NAME" is deprecated in favor of "native.repository_name()". Please rename it.`,
			`:7: Global variable "REPOSITORY_NAME" is deprecated in favor of "native.repository_name()". Please rename it.`,
		}, scopeBzl)

	checkFindings(t, "repository-name", `
REPOSITORY_NAME = "foo"
foo(a = REPOSITORY_NAME)
`, []string{}, scopeBzl)

	checkFindings(t, "repository-name", `
load(":foo.bzl", "REPOSITORY_NAME")
foo(a = REPOSITORY_NAME)
`, []string{}, scopeBzl)
}

func TestFileTypeNameWarning(t *testing.T) {
	checkFindings(t, "filetype", `
rule1(types=FileType([".cc", ".h"]))
rule2(types=FileType(types=[".cc", ".h"]))

FileType(foobar)

def macro1():
    a = FileType([".py"])

def macro2():
    FileType = foo
    b = FileType([".java"])
`, []string{
		":1: The FileType function is deprecated.",
		":2: The FileType function is deprecated.",
		":4: The FileType function is deprecated.",
		":7: The FileType function is deprecated.",
	}, scopeBzl)

	checkFindings(t, "filetype", `
FileType = foo

rule1(types=FileType([".cc", ".h"]))
rule2(types=FileType(types=[".cc", ".h"]))
`, []string{}, scopeBzl)
}

func TestOutputGroupWarning(t *testing.T) {
	checkFindingsAndFix(t, "output-group", `
def _impl(ctx):
    bin = ctx.attr.my_dep.output_group.bin
`, `
def _impl(ctx):
    bin = ctx.attr.my_dep[OutputGroupInfo].bin
`,
		[]string{
			`:2: "ctx.attr.dep.output_group" is deprecated in favor of "ctx.attr.dep[OutputGroupInfo]".`,
		},
		scopeBzl)
}

func TestNativeGitRepositoryWarning(t *testing.T) {
	checkFindingsAndFix(t, "git-repository", `
"""My file"""

def macro():
    git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def macro():
    git_repository(foo, bar)
`,
		[]string{
			`:4: Function "git_repository" is not global anymore and needs to be loaded from "@bazel_tools//tools/build_defs/repo:git.bzl".`,
		},
		scopeBzl)

	checkFindingsAndFix(t, "git-repository", `
"""My file"""

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`,
		[]string{
			`:4: Function "git_repository" is not global anymore and needs to be loaded from "@bazel_tools//tools/build_defs/repo:git.bzl".`,
			`:5: Function "new_git_repository" is not global anymore and needs to be loaded from "@bazel_tools//tools/build_defs/repo:git.bzl".`,
		},
		scopeBzl)

	checkFindingsAndFix(t, "git-repository", `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

def macro():
    git_repository(foo, bar)
    new_git_repository(foo, bar)
`,
		[]string{
			`:7: Function "new_git_repository" is not global anymore and needs to be loaded from "@bazel_tools//tools/build_defs/repo:git.bzl".`,
		},
		scopeBzl)
}

func TestNativeHttpArchiveWarning(t *testing.T) {
	checkFindingsAndFix(t, "http-archive", `
"""My file"""

def macro():
    http_archive(foo, bar)
`, `
"""My file"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def macro():
    http_archive(foo, bar)
`,
		[]string{
			`:4: Function "http_archive" is not global anymore and needs to be loaded from "@bazel_tools//tools/build_defs/repo:http.bzl".`,
		},
		scopeBzl)
}

func TestContextArgsAPIWarning(t *testing.T) {
	checkFindingsAndFix(t, "ctx-args", `
def impl(ctx):
    args = ctx.actions.args()
    args.add(foo, bar)
    args.add(foo, bar, before_each = aaa)
    args.add(foo, bar, join_with = bbb)
    args.add(foo, bar, before_each = ccc, join_with = ddd)
    args.add(foo, bar, map_fn = eee)
    args.add(foo, bar, map_fn = fff, before_each = ggg)
    args.add(foo, bar, map_fn = hhh, join_with = iii)
    args.add(foo, bar, map_fn = jjj, before_each = kkk, join_with = lll)
`, `
def impl(ctx):
    args = ctx.actions.args()
    args.add(foo, bar)
    args.add_all(foo, bar, before_each = aaa)
    args.add_joined(foo, bar, join_with = bbb)
    args.add_joined(foo, bar, format_each = ccc + "%s", join_with = ddd)
    args.add_all(foo, bar, map_each = eee)
    args.add_all(foo, bar, map_each = fff, before_each = ggg)
    args.add_joined(foo, bar, map_each = hhh, join_with = iii)
    args.add_joined(foo, bar, map_each = jjj, format_each = kkk + "%s", join_with = lll)
`,
		[]string{
			`:4: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:5: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:6: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:7: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:8: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:9: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
			`:10: "ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`,
		},
		scopeBzl)
}

func TestAttrOutputDefault(t *testing.T) {
	checkFindings(t, "attr-output-default", `
rule(
  attrs = {
      "foo": attr.output(default="foo"),
      "bar": attr.output(not_default="foo"),
      "baz": attr.string(default="foo"),
  }
)
`,
		[]string{
			`:3: The "default" parameter for attr.output() is deprecated.`,
		},
		scopeBzl)
}

func TestAttrLicense(t *testing.T) {
	checkFindings(t, "attr-license", `
rule(
  attrs = {
      "foo": attr.license(foo),
      "bar": attr.license(),
      "baz": attr.no_license(),
  }
)
`, []string{
		`:3: "attr.license()" is deprecated and shouldn't be used.`,
		`:4: "attr.license()" is deprecated and shouldn't be used.`,
	}, scopeBzl)
}

func TestRuleImplReturn(t *testing.T) {
	checkFindings(t, "rule-impl-return", `
def _impl(ctx):
  return struct()

rule(implementation=_impl)
`, []string{
		`:2: Avoid using the legacy provider syntax.`,
	}, scopeBzl)

	checkFindings(t, "rule-impl-return", `
def _impl(ctx):
  if True:
    return struct()
  return

x = rule(_impl, attrs = {})
`, []string{
		`:3: Avoid using the legacy provider syntax.`,
	}, scopeBzl)

	checkFindings(t, "rule-impl-return", `
def _impl(ctx):
  pass  # no return statements

x = rule(_impl, attrs = {})
`, []string{}, scopeBzl)

	checkFindings(t, "rule-impl-return", `
def _impl1():  # not used as a rule implementation function
  return struct()

def _impl2():  # no structs returned
  if x:
    return []
  elif y:
    return foo()
  return

x = rule(
  implementation=_impl2,
)

rule(
  _impl3,  # not defined here
)

rule(
  _impl1(),  # not an identifier
)

rule()  # no parameters
rule(foo = bar)  # no matching parameters
`, []string{}, scopeBzl)
}

func TestNativeAndroidWarning(t *testing.T) {
	checkFindingsAndFix(t, "native-android", `
"""My file"""

def macro():
    aar_import()
    android_library()
    native.android_library()
    native.android_local_test()

android_binary()
`, fmt.Sprintf(`
"""My file"""

load(%q, "aar_import", "android_binary", "android_library", "android_local_test")

def macro():
    aar_import()
    android_library()
    android_library()
    android_local_test()

android_binary()
`, tables.AndroidLoadPath),
		[]string{
			fmt.Sprintf(`:4: Function "aar_import" is not global anymore and needs to be loaded from "%s".`, tables.AndroidLoadPath),
			fmt.Sprintf(`:5: Function "android_library" is not global anymore and needs to be loaded from "%s".`, tables.AndroidLoadPath),
			fmt.Sprintf(`:6: Function "android_library" is not global anymore and needs to be loaded from "%s".`, tables.AndroidLoadPath),
			fmt.Sprintf(`:7: Function "android_local_test" is not global anymore and needs to be loaded from "%s".`, tables.AndroidLoadPath),
			fmt.Sprintf(`:9: Function "android_binary" is not global anymore and needs to be loaded from "%s".`, tables.AndroidLoadPath),
		},
		scopeBzl|scopeBuild)
}

func TestNativeCcWarning(t *testing.T) {
	checkFindingsAndFix(t, "native-cc", `
"""My file"""

def macro():
    cc_library()
    native.cc_binary()
    cc_test()
    cc_proto_library()
    native.fdo_prefetch_hints()
    native.objc_library()
    objc_import()
    cc_toolchain()
    native.cc_toolchain_suite()

fdo_profile()
cc_import()
`, fmt.Sprintf(`
"""My file"""

load(%q, "cc_binary", "cc_import", "cc_library", "cc_proto_library", "cc_test", "cc_toolchain", "cc_toolchain_suite", "fdo_prefetch_hints", "fdo_profile", "objc_import", "objc_library")

def macro():
    cc_library()
    cc_binary()
    cc_test()
    cc_proto_library()
    fdo_prefetch_hints()
    objc_library()
    objc_import()
    cc_toolchain()
    cc_toolchain_suite()

fdo_profile()
cc_import()
`, tables.CcLoadPath),
		[]string{
			fmt.Sprintf(`:4: Function "cc_library" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:5: Function "cc_binary" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:6: Function "cc_test" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:7: Function "cc_proto_library" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:8: Function "fdo_prefetch_hints" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:9: Function "objc_library" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:10: Function "objc_import" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:11: Function "cc_toolchain" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:12: Function "cc_toolchain_suite" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:14: Function "fdo_profile" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
			fmt.Sprintf(`:15: Function "cc_import" is not global anymore and needs to be loaded from "%s".`, tables.CcLoadPath),
		},
		scopeBzl|scopeBuild)
}

func TestNativeJavaWarning(t *testing.T) {
	checkFindingsAndFix(t, "native-java", `
"""My file"""

def macro():
    java_import()
    java_library()
    native.java_library()
    native.java_binary()

java_test()
`, fmt.Sprintf(`
"""My file"""

load(%q, "java_binary", "java_import", "java_library", "java_test")

def macro():
    java_import()
    java_library()
    java_library()
    java_binary()

java_test()
`, tables.JavaLoadPath),
		[]string{
			fmt.Sprintf(`:4: Function "java_import" is not global anymore and needs to be loaded from "%s".`, tables.JavaLoadPath),
			fmt.Sprintf(`:5: Function "java_library" is not global anymore and needs to be loaded from "%s".`, tables.JavaLoadPath),
			fmt.Sprintf(`:6: Function "java_library" is not global anymore and needs to be loaded from "%s".`, tables.JavaLoadPath),
			fmt.Sprintf(`:7: Function "java_binary" is not global anymore and needs to be loaded from "%s".`, tables.JavaLoadPath),
			fmt.Sprintf(`:9: Function "java_test" is not global anymore and needs to be loaded from "%s".`, tables.JavaLoadPath),
		},
		scopeBzl|scopeBuild)
}

func TestNativePyWarning(t *testing.T) {
	checkFindingsAndFix(t, "native-py", `
"""My file"""

def macro():
    py_library()
    py_binary()
    native.py_test()
    native.py_runtime()

py_test()
`, fmt.Sprintf(`
"""My file"""

load(%q, "py_binary", "py_library", "py_runtime", "py_test")

def macro():
    py_library()
    py_binary()
    py_test()
    py_runtime()

py_test()
`, tables.PyLoadPath),
		[]string{
			fmt.Sprintf(`:4: Function "py_library" is not global anymore and needs to be loaded from "%s".`, tables.PyLoadPath),
			fmt.Sprintf(`:5: Function "py_binary" is not global anymore and needs to be loaded from "%s".`, tables.PyLoadPath),
			fmt.Sprintf(`:6: Function "py_test" is not global anymore and needs to be loaded from "%s".`, tables.PyLoadPath),
			fmt.Sprintf(`:7: Function "py_runtime" is not global anymore and needs to be loaded from "%s".`, tables.PyLoadPath),
			fmt.Sprintf(`:9: Function "py_test" is not global anymore and needs to be loaded from "%s".`, tables.PyLoadPath),
		},
		scopeBzl|scopeBuild)
}

func TestNativeProtoWarning(t *testing.T) {
	checkFindingsAndFix(t, "native-proto", `
"""My file"""

def macro():
    proto_library()
    proto_lang_toolchain()
    native.proto_lang_toolchain()
    native.proto_library()

    ProtoInfo
    proto_common
`, fmt.Sprintf(`
"""My file"""

load(%q, "ProtoInfo", "proto_common", "proto_lang_toolchain", "proto_library")

def macro():
    proto_library()
    proto_lang_toolchain()
    proto_lang_toolchain()
    proto_library()

    ProtoInfo
    proto_common
`, tables.ProtoLoadPath),
		[]string{
			fmt.Sprintf(`:4: Function "proto_library" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
			fmt.Sprintf(`:5: Function "proto_lang_toolchain" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
			fmt.Sprintf(`:6: Function "proto_lang_toolchain" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
			fmt.Sprintf(`:7: Function "proto_library" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
			fmt.Sprintf(`:9: Symbol "ProtoInfo" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
			fmt.Sprintf(`:10: Symbol "proto_common" is not global anymore and needs to be loaded from "%s".`, tables.ProtoLoadPath),
		},
		scopeBzl|scopeBuild)
}

func TestKeywordParameters(t *testing.T) {
	checkFindingsAndFix(t, "keyword-positional-params", `
foo(key = value)
all(elements = [True, False])
any(elements = [True, False])
tuple(x = [1, 2, 3])
list(x = [1, 2, 3])
len(x = [1, 2, 3])
str(x = foo)
repr(x = foo)
bool(x = 3)
int(x = "3")
int(x = "13", base = 8)
dir(x = foo)
type(x = foo)
select(x = {})
`, `
foo(key = value)
all([True, False])
any([True, False])
tuple([1, 2, 3])
list([1, 2, 3])
len([1, 2, 3])
str(foo)
repr(foo)
bool(3)
int("3")
int("13", base = 8)
dir(foo)
type(foo)
select({})
`, []string{
		`:2: Keyword parameter "elements" for "all" should be positional.`,
		`:3: Keyword parameter "elements" for "any" should be positional.`,
		`:4: Keyword parameter "x" for "tuple" should be positional.`,
		`:5: Keyword parameter "x" for "list" should be positional.`,
		`:6: Keyword parameter "x" for "len" should be positional.`,
		`:7: Keyword parameter "x" for "str" should be positional.`,
		`:8: Keyword parameter "x" for "repr" should be positional.`,
		`:9: Keyword parameter "x" for "bool" should be positional.`,
		`:10: Keyword parameter "x" for "int" should be positional.`,
		`:11: Keyword parameter "x" for "int" should be positional.`,
		`:12: Keyword parameter "x" for "dir" should be positional.`,
		`:13: Keyword parameter "x" for "type" should be positional.`,
		`:14: Keyword parameter "x" for "select" should be positional.`,
	}, scopeEverywhere)

	checkFindingsAndFix(t, "keyword-positional-params", `
hasattr(
  x = foo,
  name = "bar",
)
getattr(
  x = foo,
  name = "bar",
)
getattr(
  x = foo,
  name = "bar",
  default = "baz",
)
`, `
hasattr(
  foo,
  "bar",
)
getattr(
  foo,
  "bar",
)
getattr(
  foo,
  "bar",
  "baz",
)
`, []string{
		`:2: Keyword parameter "x" for "hasattr" should be positional.`,
		`:3: Keyword parameter "name" for "hasattr" should be positional.`,
		`:6: Keyword parameter "x" for "getattr" should be positional.`,
		`:7: Keyword parameter "name" for "getattr" should be positional.`,
		`:10: Keyword parameter "x" for "getattr" should be positional.`,
		`:11: Keyword parameter "name" for "getattr" should be positional.`,
		`:12: Keyword parameter "default" for "getattr" should be positional.`,
	}, scopeEverywhere)
}

func TestProvider(t *testing.T) {
	checkFindings(t, "provider-params", `provider(doc = "doc", fields = [])`, []string{}, scopeBzl)
	checkFindings(t, "provider-params", `provider("doc", fields = [])`, []string{}, scopeBzl)
	checkFindings(t, "provider-params", `provider(fields = None, doc = "doc")`, []string{}, scopeBzl)

	checkFindings(t, "provider-params", `provider(fields = [])`,
		[]string{`1: Calls to 'provider' should provide a documentation`}, scopeBzl)
	checkFindings(t, "provider-params", `provider(doc = "doc")`,
		[]string{`1: Calls to 'provider' should provide a list of fields:`}, scopeBzl)
	checkFindings(t, "provider-params", `p = provider()`,
		[]string{`1: Calls to 'provider' should provide a list of fields and a documentation:`}, scopeBzl)
}
