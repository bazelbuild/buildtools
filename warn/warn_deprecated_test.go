package warn

import "testing"

func TestDeprecatedFunction(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
def foo():
  """
  This is a function foo.

  Please use it in favor of the
  deprecated function bar
  """
  pass

def bar():
  """
  This is a function bar.

  Deprecated:
    please use foo instead.
  """
  pass
`,
		"test/package/invalid.bzl": `
This is not a valid Starlark file
`,
	})()

	checkFindings(t, "deprecated-function", `
load(":foo.bzl", "foo", "bar", "baz")
load("//test/package:foo.bzl", "foo", "bar", "baz")
load(":invalid.bzl", "foo", "bar", "baz")
load(":nonexistent.bzl", "foo", "bar", "baz")
`,
		[]string{
			`1: The function "bar" defined in "//test/package/foo.bzl" is deprecated.`,
			`2: The function "bar" defined in "//test/package/foo.bzl" is deprecated.`,
		},
		scopeEverywhere)
}

func TestDeprecatedFunctionNoReader(t *testing.T) {
	checkFindings(t, "deprecated-function", `
load(":foo.bzl", "foo", "bar", "baz")
load("//test/package:foo.bzl", "foo", "bar", "baz")
load(":invalid.bzl", "foo", "bar", "baz")
load(":nonexistent.bzl", "foo", "bar", "baz")
`,
		[]string{},
		scopeEverywhere)
}
