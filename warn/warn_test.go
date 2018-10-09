package warn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/testutils"
)

func getFilename(isBuildFile bool) string {
	if isBuildFile {
		return "BUILD"
	}
	return "test_file.bzl"
}

func getFindings(category, input string, isBuildFile bool) []*Finding {
	input = strings.TrimLeft(input, "\n")
	buildFile, err := build.Parse(getFilename(isBuildFile), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return FileWarnings(buildFile, "the_package", []string{category}, false)
}

func compareFinding(t *testing.T, input string, expected []string, findings []*Finding) {
	// We ensure that there is the expected number of warnings.
	// At the moment, we check only the line numbers.
	if len(expected) != len(findings) {
		t.Errorf("Input: %s", input)
		t.Errorf("number of matches: %d, want %d", len(findings), len(expected))
		t.Errorf("expected findings: %v", expected)
		for _, f := range findings {
			t.Errorf("got: %d: %s", f.Start.Line, f.Message)
		}
		return
	}
	for i := range findings {
		msg := fmt.Sprintf(":%d: %s", findings[i].Start.Line, findings[i].Message)
		if !strings.Contains(msg, expected[i]) {
			t.Errorf("Input: %s", input)
			t.Errorf("got:  `%s`,\nwant: `%s`", msg, expected[i])
		}
	}
}

func checkFix(t *testing.T, category, input, expected string, isBuildFile bool) {
	buildFile, err := build.Parse(getFilename(isBuildFile), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	goldenFile, err := build.Parse(getFilename(isBuildFile), []byte(expected))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	FixWarnings(buildFile, "the_package", []string{category})
	have := build.Format(buildFile)
	want := build.Format(goldenFile)
	if !bytes.Equal(have, want) {
		fileType := "bzl"
		if isBuildFile {
			fileType = "BUILD"
		}
		t.Errorf("fixed %s (type %s) incorrectly: diff shows -expected, +ours", input, fileType)
		testutils.Tdiff(t, want, have)
	}
}

func checkFindings(t *testing.T, category, input string, expected []string, isBuildFileSpecific bool) {
	// The same as checkFindingsAndFix but ensure that fixes don't change the file (except for formatting)
	checkFindingsAndFix(t, category, input, input, expected, isBuildFileSpecific)
}

func checkFindingsAndFix(t *testing.T, category, input, output string, expected []string, isBuildFileSpecific bool) {
	// All warnings should be found for BUILD-files
	compareFinding(t, input, expected, getFindings(category, input, true))
	checkFix(t, category, input, output, true)

	if isBuildFileSpecific {
		// BUILD-file specific warnings shouldn't be shown for .bzl files
		expected = []string{}
		// BUILD-file specific fixes shouldn't affect .bzl files
		output = input
	}
	compareFinding(t, input, expected, getFindings(category, input, false))
	checkFix(t, category, input, output, false)
}

func TestNoEffect(t *testing.T) {
	checkFindings(t, "no-effect", `
"""Docstring."""
def bar():
    """Other Docstring"""
    fct()
    pass
    return 2

[f() for i in range(3)] # top-level comprehension is okay
`,
		[]string{},
		false)

	checkFindings(t, "no-effect", `
def foo():
    [fct() for i in range(3)]
	`,
		[]string{":2: Expression result is not used. Use a for-loop instead"},
		false)

	checkFindings(t, "no-effect", `None`,
		[]string{":1: Expression result is not used."},
		false)

	checkFindings(t, "no-effect", `
foo             # 1
foo()

def bar():
    [1, 2]      # 5
    if True:
      "string"  # 7
`,
		[]string{":1:", ":5:", ":7:"},
		false)

	checkFindings(t, "no-effect", `
# A comment

"""A docstring"""

# Another comment

"""Not a docstring"""

def bar():
    """A docstring"""
    foo
    return foo
`,
		[]string{":7:", ":11:"},
		false)

	checkFindings(t, "no-effect", `
foo == bar
foo = bar
a + b
c // d
-e
foo != bar

foo += bar
bar -= bar

`,
		[]string{":1:", ":3:", ":4:", ":5:", ":6:"},
		false)
}

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
		false)
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
		true)
}

func TestWarnUnusedLoad(t *testing.T) {
	checkFindingsAndFix(t, "unused-load", `
load(":f.bzl", "s1", "s2")
load("f", "s1")
foo(name = s1)`, `
load(":f.bzl", "s1")
foo(name = s1)`,
		[]string{":1: Loaded symbol \"s2\" is unused.",
			":2: Symbol \"s1\" has already been loaded."},
		false)

	checkFindingsAndFix(t, "unused-load", `
load(
  ":f.bzl",
   "s1",
   "s2",  # @unused (s2)
)

# @unused - both s3 and s4
load(
  ":f.bzl",
   "s3",
   "s4",
)`, `
load(
  ":f.bzl",
   "s2",  # @unused (s2)
)

# @unused - both s3 and s4
load(
  ":f.bzl",
   "s3",
   "s4",
)`,
		[]string{":3: Loaded symbol \"s1\" is unused."},
		false)

	checkFindingsAndFix(t, "unused-load", `
load(":f.bzl", "x")
x = "unused"`, `
x = "unused"`,
		[]string{":1: Loaded symbol \"x\" is unused."},
		false)
}

func TestWarnUnusedVariables(t *testing.T) {
	checkFindings(t, "unused-variable", `
load(":f.bzl", "x")
x = "unused"
y = "also unused"
z = "name"
cc_library(name = z)`,
		[]string{":2: Variable \"x\" is unused.",
			":3: Variable \"y\" is unused."},
		true)

	checkFindings(t, "unused-variable", `
a = 1
b = 2
c = 3
d = (a if b else c)  # only d is unused
e = 5 # @unused
# @unused
f = 7`,
		[]string{":4: Variable \"d\" is unused."},
		true)

	checkFindings(t, "unused-variable", `
a = 1

def foo():
  b = 2
  c = 3
  d = (a if b else c)  # only d is unused
  e = 5 # @unused
  # @unused
  f = 7
  g = 8
  return g`,
		[]string{":6: Variable \"d\" is unused."},
		true)

	checkFindings(t, "unused-variable", `
a = 1

def foo(c):
  b = 2
	return c

def bar(b):
  c = 3
	print(b)`,
		[]string{
			":1: Variable \"a\" is unused.",
			":4: Variable \"b\" is unused.",
			":8: Variable \"c\" is unused.",
		},
		true)
}

func TestRedefinedVariable(t *testing.T) {
	checkFindings(t, "redefined-variable", `
x = "old_value"
x = "new_value"
cc_library(name = x)`,
		[]string{":2: Variable \"x\" has already been defined."},
		false)

	checkFindings(t, "redefined-variable", `
x = "a"

def foo():
  x = "b"
  y = "c"
  y = "d"

def bar():
  x = "e"
  y = "f"
  y = "g"`,
		[]string{},
		false)
}

func TestPackageOnTop(t *testing.T) {
	checkFindings(t, "package-on-top", `
my_macro(name = "foo")
package()`,
		[]string{":2: Package declaration should be at the top of the file, after the load() statements, but before any call to a rule or a macro. package_group() and licenses() may be called before package()."},
		false)
}

func TestPositionalArguments(t *testing.T) {
	checkFindings(t, "positional-args", `
my_macro(foo = "bar")
my_macro("foo", "bar")`,
		[]string{":2: All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax."},
		true)
}

func TestIntegerDivision(t *testing.T) {
	checkFindingsAndFix(t, "integer-division", `
a = b / c
d /= e
`, `
a = b // c
d //= e
`,
		[]string{
			":1: The \"/\" operator for integer division is deprecated in favor of \"//\".",
			":2: The \"/=\" operator for integer division is deprecated in favor of \"//=\".",
		},
		false)
}
