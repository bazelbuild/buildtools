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

import "testing"

func TestConstantGlob(t *testing.T) {
	checkFindings(t, "constant-glob", `
cc_library(srcs = glob(["foo.cc"]))
cc_library(srcs = glob(include = ["foo.cc"]))
cc_library(srcs = glob(include = ["foo.cc"], exclude = ["bar.cc"]))
cc_library(srcs = glob(exclude = ["bar.cc"], include = ["foo.cc"]))
cc_library(srcs =
	["constant"] + glob([
		"*.cc",
		"test.cpp",
		])
	)
cc_library(srcs = glob(["*.cc"]))
cc_library(srcs = glob(["*.cc"], exclude = ["bar.cc"]))
cc_library(srcs = glob(include = ["*.cc"], exclude = ["bar.cc"]))
cc_library(srcs = glob(exclude = ["bar.cc"], include = ["*.cc"]))`,
		[]string{`:1: Glob pattern "foo.cc" has no wildcard`,
			`:2: Glob pattern "foo.cc" has no wildcard`,
			`:3: Glob pattern "foo.cc" has no wildcard`,
			`:4: Glob pattern "foo.cc" has no wildcard`,
			`:8: Glob pattern "test.cpp" has no wildcard`},
		scopeBuild|scopeBzl|scopeWorkspace)
}

func TestNativeInBuildFiles(t *testing.T) {
	checkFindingsAndFix(t, "native-build", `
native.package("foo")

native.cc_library(name = "lib")
`, `
package("foo")

cc_library(name = "lib")
`, []string{
		`:1: The "native" module shouldn't be used in BUILD files, its members are available as global symbols.`,
		`:3: The "native" module shouldn't be used in BUILD files, its members are available as global symbols.`,
	}, scopeBuild)
}

func TestNativePackage(t *testing.T) {
	checkFindings(t, "native-package", `
native.package("foo")

native.cc_library(name = "lib")
`, []string{
		`:1: "native.package()" shouldn't be used in .bzl files.`,
	}, scopeBzl)
}

func TestDuplicatedName(t *testing.T) {
	checkFindings(t, "duplicated-name", `
cc_library(name = "x")
cc_library(name = "y")
py_library(name = "x")
py_library(name = "z")
php_library(name = "x")`,
		[]string{
			`:3: A rule with name "x" was already found on line 1`,
			`:5: A rule with name "x" was already found on line 1`,
		}, scopeBuild|scopeWorkspace)

	checkFindings(t, "duplicated-name", `
exports_files(["foo.txt"])
[macro(name = "bar_%s" % i) for i in ii]
`,
		[]string{},
		scopeBuild|scopeWorkspace)
}

func TestPositionalArgumentsDoesNotWarnForNamedArguments(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro = macro()
my_rule = rule()
def my_function(foo):
  pass

my_macro(foo = "bar")
my_rule(foo = "bar")
my_function(foo = "bar")
`,
		[]string{},
		scopeBuild)
}

func TestPositionalArgumentsDoesNotWarnForAllowlistedFunctions(t *testing.T) {
	checkFindings(t, "positional-args", `
register_toolchains(
	"//foo",
	"//bar",
)`,
		[]string{},
		scopeBuild)
}

func TestPositionalArgumentsWarnsForPositionalMacrosOrRuleCalls(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro = macro()
my_rule = rule()
def my_function(foo):
  pass

my_macro("foo", "bar")
my_rule("foo", "bar")
my_function("foo", "bar")
`,
		[]string{
			`6: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_macro" with positional arguments.`,
			`7: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_rule" with positional arguments.`,
		},
		scopeBuild)
}

func TestPositionalArgumentsWarnsWhenCalledInNestedContexts(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro = macro()
my_rule = rule()
def my_function(foo):
  pass

other_function(foo = my_macro(x))
other_function(foo = my_rule(x))
other_function(foo = my_function(x))

[my_macro(foo) for foo in bar]
[my_rule(foo) for foo in bar]
[my_function(foo) for foo in bar]
`,
		[]string{
			`6: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_macro" with positional arguments.`,
			`7: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_rule" with positional arguments.`,
			`10: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_macro" with positional arguments.`,
			`11: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro "my_rule" with positional arguments.`,
		},
		scopeBuild)

}

func TestKwargsInBuildFilesWarning(t *testing.T) {
	checkFindings(t, "build-args-kwargs", `
cc_library(
  name = "foo",
  *args,
  **kwargs,
)

foo(*bar(**kgs))`,
		[]string{
			":3: *args are not allowed in BUILD files.",
			":4: **kwargs are not allowed in BUILD files.",
			":7: *args are not allowed in BUILD files.",
			":7: **kwargs are not allowed in BUILD files.",
		},
		scopeBuild)

	checkFindings(t, "build-args-kwargs", `
cc_library(
  name = "foo",
  -args,
)

foo(not bar(-kgs))`,
		[]string{},
		scopeBuild)
}

func TestPrintWarning(t *testing.T) {
	checkFindings(t, "print", `
foo()

print("foo")

def f(x):
  print(x)

  g(x) or print("not g")
`,
		[]string{
			`:3: "print()" is a debug function and shouldn't be submitted.`,
			`:6: "print()" is a debug function and shouldn't be submitted.`,
			`:8: "print()" is a debug function and shouldn't be submitted.`,
		},
		scopeBazel)
}

func TestMakeLocationVariable(t *testing.T) {
	checkFindings(t, "make-location", `
genrule(
    name = "a",
    srcs = [":foo"],
    outs = ["out"],
    cmd = "cp $(location :foo) $@",
)

genrule(
    name = "b",
    srcs = [":a", ":b"],
    outs = ["out"],
    cmd = "cat $(locations :a :b) > $@",
)

genrule(
    name = "c",
    srcs = [":foo"],
    outs = ["out"],
    cmd = "cp $$(location :foo) $$@",
)

cc_test(
    name = "d",
    args = ["--config=$(execpath :cfg)"],
)

load("location", "symbol")
`,
		[]string{
			`:5: The "$(location)" and "$(locations)" make variables are deprecated. Use "$(execpath ...)" or "$(rootpath ...)" instead.`,
			`:12: The "$(location)" and "$(locations)" make variables are deprecated. Use "$(execpath ...)" or "$(rootpath ...)" instead.`,
			`:19: The "$(location)" and "$(locations)" make variables are deprecated. Use "$(execpath ...)" or "$(rootpath ...)" instead.`,
		},
		scopeBuild)
}

func TestPathMapping(t *testing.T) {
	checkFindings(t, "path-mapping", `
load("@bazel_skylib//lib:paths.bzl", "paths")

def _dirname(file):
    return file.dirname

def _add_srcs(args, srcs):
    args.add_all([src.path for src in srcs])

def _impl(ctx):
    args = ctx.actions.args()
    actions = ctx.actions

    # Compatible with path mapping
    args.add(src)
    args.add("--out", out)
    args.add_all(dirs, expand_directories = False)
    args.add_all([src], map_each = _dirname)
    args.add_all(srcs, format_each = "--src=%s")
    args.add(src.short_path)
    args.add(src.basename)
    args.add(paths.dirname(src.short_path))
    ctx.actions.run(
        outputs = [out],
        executable = ctx.executable._tool,
        arguments = [args],
    )
    ctx.actions.run_shell(
        outputs = [out],
        command = "cp $1 $2",
        arguments = [args],
    )

    # Incompatible with path mapping
    args.add(src.path)
    args.add("--out", out.path)
    args.add(src.dirname)
    args.add(dir.path)
    args.add_all([src.path for src in srcs])
    args.add_joined([src.path for src in srcs], join_with = ",")
    args.add("--bin=%s" % ctx.bin_dir.path)
    args.add("--gen={}".format(ctx.genfiles_dir.path))
    args.add("--root=" + src.root.path)
    args.add(paths.join(ctx.bin_dir.path, "include"))
    args.add(src.path if src else out.path)
    ctx.actions.run(
        outputs = [out],
        executable = tool.path,
        arguments = [src.path, "--out", out.path],
    )
    ctx.actions.run_shell(
        outputs = [out],
        command = "cp %s %s" % (src.path, out.path),
    )
    actions.run_shell(
        outputs = [out],
        command = "cp $1 " + out.dirname,
        arguments = [src.path],
    )

    # Unrelated
    other.add(src.path)
    ctx.actions.other(arguments = [src.path])
    ctx.actions.run(
        outputs = [out],
        executable = ctx.executable._tool,
        arguments = [args],
        progress_message = "Building %s" % out.path,
    )
`,
		[]string{
			`:7: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:34: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:35: "out.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:36: "src.dirname" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:37: "dir.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:38: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:39: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:40: "ctx.bin_dir.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:41: "ctx.genfiles_dir.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:42: "src.root.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:43: "ctx.bin_dir.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:44: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:44: "out.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:47: "tool.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:48: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:48: "out.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:52: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:52: "out.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:56: "out.dirname" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
			`:57: "src.path" is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
		},
		scopeBzl)
}
