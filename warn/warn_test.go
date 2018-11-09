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
	checkFindingsAndFix(t, "load", `
load(":f.bzl", "s1", "s2")
load("f", "s1")
foo(name = s1)`, `
load(":f.bzl", "s1")
foo(name = s1)`,
		[]string{":1: Loaded symbol \"s2\" is unused.",
			":2: Symbol \"s1\" has already been loaded."},
		false)

	checkFindingsAndFix(t, "load", `
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

	checkFindingsAndFix(t, "load", `
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

func TestLoadOnTop(t *testing.T) {
	checkFindingsAndFix(t, "load-on-top", `
foo()
load(":f.bzl", "x")
x()`, `
load(":f.bzl", "x")

foo()

x()`,
		[]string{":2: Load statements should be at the top of the file."},
		false)

	checkFindingsAndFix(t, "load-on-top", `
"""Docstring"""

# Comment block

# this is foo
foo()

# load
load(":f.bzl", "bar")

# this is bar
bar()

# another load
load(":f.bzl", "foobar")`, `
"""Docstring"""

# Comment block

# load
load(":f.bzl", "bar")

# another load
load(":f.bzl", "foobar")

# this is foo
foo()

# this is bar
bar()`,
		[]string{
			":9: Load statements should be at the top of the file.",
			":15: Load statements should be at the top of the file.",
		}, false)
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

func TestDictionaryConcatenation(t *testing.T) {
	checkFindings(t, "dict-concatenation", `
d = {}

d + foo
foo + d
d + foo + bar  # Should trigger 2 warnings: (d + foo) is recognized as a dict
foo + bar + d  # Should trigger 1 warning: (foo + bar) is unknown
d += foo + bar
`,
		[]string{
			":3: Dictionary concatenation is deprecated.",
			":4: Dictionary concatenation is deprecated.",
			":5: Dictionary concatenation is deprecated.",
			":5: Dictionary concatenation is deprecated.",
			":6: Dictionary concatenation is deprecated.",
			":7: Dictionary concatenation is deprecated.",
		},
		false)
}

func TestStringIteration(t *testing.T) {
	checkFindings(t, "string-iteration", `
s = "foo" + bar

max(s)
min(s)
all(s)
any(s)
reversed(s)
zip(s, a, b)
zip(a, s)

[foo(x) for x in s]

for x in s:
    pass

# The following iterations over a list don't trigger warnings

l = list()

max(l)
zip(l, foo)
[foo(x) for x in l]

for x in l:
    pass
`,
		[]string{
			":3: String iteration is deprecated.",
			":4: String iteration is deprecated.",
			":5: String iteration is deprecated.",
			":6: String iteration is deprecated.",
			":7: String iteration is deprecated.",
			":8: String iteration is deprecated.",
			":9: String iteration is deprecated.",
			":11: String iteration is deprecated.",
			":13: String iteration is deprecated.",
		},
		false)
}

func TestDepsetIteration(t *testing.T) {
	checkFindingsAndFix(t, "depset-iteration", `
d = depset([1, 2, 3]) + bar

max(d + foo)
min(d)
all(d)
any(d)
sorted(d)
zip(
    d,
    a,
    b,
)
zip(
     a,
     d,
)
list(d)
tuple(d)
depset(d)
len(d)
1 in d
2 not in d

[foo(x) for x in d]

for x in d:
    pass

# Non-iteration is ok

foobar(d)
d == b

# The following iterations over a list don't trigger warnings

l = list([1, 2, 3])

max(l)
zip(l, foo)
[foo(x) for x in l]
1 in l

for x in l:
    pass
`, `
d = depset([1, 2, 3]) + bar

max((d + foo).to_list())
min(d.to_list())
all(d.to_list())
any(d.to_list())
sorted(d.to_list())
zip(
    d.to_list(),
    a,
    b,
)
zip(
    a,
    d.to_list(),
)
d.to_list()
tuple(d.to_list())
depset(d.to_list())
len(d.to_list())
1 in d.to_list()
2 not in d.to_list()

[foo(x) for x in d.to_list()]

for x in d.to_list():
    pass

# Non-iteration is ok

foobar(d)
d == b

# The following iterations over a list don't trigger warnings

l = list([1, 2, 3])

max(l)
zip(l, foo)
[foo(x) for x in l]
1 in l

for x in l:
    pass
`,
		[]string{
			":3: Depset iteration is deprecated.",
			":4: Depset iteration is deprecated.",
			":5: Depset iteration is deprecated.",
			":6: Depset iteration is deprecated.",
			":7: Depset iteration is deprecated.",
			":9: Depset iteration is deprecated.",
			":15: Depset iteration is deprecated.",
			":17: Depset iteration is deprecated.",
			":18: Depset iteration is deprecated.",
			":19: Depset iteration is deprecated.",
			":20: Depset iteration is deprecated.",
			":21: Depset iteration is deprecated.",
			":22: Depset iteration is deprecated.",
			":24: Depset iteration is deprecated.",
			":26: Depset iteration is deprecated.",
		},
		false)
}
