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
		[]string{":1: Glob pattern `foo.cc` has no wildcard",
			":6: Glob pattern `test.cpp` has no wildcard"},
		scopeEverywhere)
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
		[]string{":3: A rule with name `x' was already found on line 1",
			":5: A rule with name `x' was already found on line 1"},
		scopeBuild|scopeWorkspace)
}

func TestPositionalArguments(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro(foo = "bar")
my_macro("foo", "bar")`,
		[]string{":2: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax."},
		scopeBuild|scopeWorkspace)

	checkFindings(t, "positional-args", `
register_toolchains(
	"//foo",
	"//bar",
)`,
		[]string{},
		scopeBuild|scopeWorkspace)
}
