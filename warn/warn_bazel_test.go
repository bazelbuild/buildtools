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
cc_library(srcs = glob(["*.cc"]))
cc_library(srcs =
  ["constant"] + glob([
    "*.cc",
    "test.cpp",
  ])
)`,
		[]string{`:1: Glob pattern "foo.cc" has no wildcard`,
			`:6: Glob pattern "test.cpp" has no wildcard`},
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

func TestPositionalArguments(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro(foo = "bar")
my_macro("foo", "bar")
my_macro(foo = bar(x))
[my_macro(foo) for foo in bar]`,
		[]string{
			":2: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.",
			":4: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.",
		},
		scopeBuild)

	checkFindings(t, "positional-args", `
register_toolchains(
	"//foo",
	"//bar",
)`,
		[]string{},
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
