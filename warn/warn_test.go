package warn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/testutils"
)

const (
	scopeBuild      = build.TypeBuild
	scopeBzl        = build.TypeDefault
	scopeWorkspace  = build.TypeWorkspace
	scopeEverywhere = scopeBuild | scopeBzl | scopeWorkspace
)

func getFilename(fileType build.FileType) string {
	switch fileType {
	case build.TypeBuild:
		return "BUILD"
	case build.TypeWorkspace:
		return "WORKSPACE"
	default:
		return "test_file.bzl"
	}
}

func getFindings(category, input string, fileType build.FileType) []*Finding {
	input = strings.TrimLeft(input, "\n")
	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return FileWarnings(buildFile, "the_package", []string{category}, false)
}

func compareFindings(t *testing.T, category, input string, expected []string, scope, fileType build.FileType) {
	// If scope doesn't match the file type, no warnings are expected
	if scope&fileType == 0 {
		expected = []string{}
	}

	findings := getFindings(category, input, fileType)
	// We ensure that there is the expected number of warnings.
	// At the moment, we check only the line numbers.
	if len(expected) != len(findings) {
		t.Errorf("Input: %s", input)
		t.Errorf("number of matches: %d, want %d", len(findings), len(expected))
		for _, e := range expected {
			t.Errorf("expected: %s", e)
		}
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

func checkFix(t *testing.T, category, input, expected string, scope, fileType build.FileType) {
	// If scope doesn't match the file type, no changes are expected
	if scope&fileType == 0 {
		expected = input
	}

	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	goldenFile, err := build.Parse(getFilename(fileType), []byte(expected))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	FixWarnings(buildFile, "the_package", []string{category}, false)
	have := build.Format(buildFile)
	want := build.Format(goldenFile)
	if !bytes.Equal(have, want) {
		t.Errorf("fixed a test (type %s) incorrectly:\ninput:\n%s\ndiff (-expected, +ours)\n",
			fileType, input)
		testutils.Tdiff(t, want, have)
	}
}

func checkFindings(t *testing.T, category, input string, expected []string, scope build.FileType) {
	// The same as checkFindingsAndFix but ensure that fixes don't change the file (except for formatting)
	checkFindingsAndFix(t, category, input, input, expected, scope)
}

func checkFindingsAndFix(t *testing.T, category, input, output string, expected []string, scope build.FileType) {
	fileTypes := []build.FileType{
		build.TypeDefault,
		build.TypeBuild,
		build.TypeWorkspace,
	}

	for _, fileType := range fileTypes {
		compareFindings(t, category, input, expected, scope, fileType)
		checkFix(t, category, input, output, scope, fileType)
		checkFix(t, category, output, output, scope, fileType)
	}
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
		scopeEverywhere)

	checkFindings(t, "no-effect", `
def foo():
    [fct() for i in range(3)]
	`,
		[]string{":2: Expression result is not used. Use a for-loop instead"},
		scopeEverywhere)

	checkFindings(t, "no-effect", `None`,
		[]string{":1: Expression result is not used."},
		scopeEverywhere)

	checkFindings(t, "no-effect", `
foo             # 1
foo()

def bar():
    [1, 2]      # 5
    if True:
      "string"  # 7
`,
		[]string{":1:", ":5:", ":7:"},
		scopeEverywhere)

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
		scopeEverywhere)

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
		scopeEverywhere)
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
		scopeEverywhere)
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

func TestWarnUnusedLoad(t *testing.T) {
	checkFindingsAndFix(t, "load", `
load(":f.bzl", "s1", "s2")
load(":bar.bzl", "s1")
foo(name = s1)`, `
load(":bar.bzl", "s1")
foo(name = s1)`,
		[]string{
			":1: Symbol \"s1\" has already been loaded.",
			":1: Loaded symbol \"s2\" is unused."},
		scopeEverywhere)

	checkFindingsAndFix(t, "load", `
load("foo", "a", "b", "c")
load("bar", "a", "d", "e")

z = a + b + d`, `
load("foo", "b")
load("bar", "a", "d")

z = a + b + d`,
		[]string{
			":2: Loaded symbol \"e\" is unused.",
			":1: Symbol \"a\" has already been loaded.",
			":1: Loaded symbol \"c\" is unused.",
		},
		scopeEverywhere)

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
		scopeEverywhere)

	checkFindingsAndFix(t, "load", `
load(":f.bzl", "x")
x = "unused"`, `
x = "unused"`,
		[]string{":1: Loaded symbol \"x\" is unused."},
		scopeEverywhere)
}

func TestWarnSameOriginLoad(t *testing.T) {
	category := "same-origin-load"

	checkFindingsAndFix(t, category, `
	load(
		":f.bzl",
		"s1"
	)
	load(":t.bzl", "s3")
	load(
		":f.bzl",
		"s2"
	)`, `
	load(
		":f.bzl",
		"s1",
		"s2"
	)
	load(":t.bzl", "s3")`,
		[]string{":7: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one."},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	load(
		":f.bzl",
		"s1"
	)
	load(
		":f.bzl",
		"s2"
	)
	load(
		":f.bzl",
		"s3"
	)`, `
	load(
		":f.bzl",
		"s1",
		"s2",
		"s3"
	)`,
		[]string{":6: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one.",
			":10: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one."},
		scopeEverywhere,
	)
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
		scopeBuild|scopeWorkspace)

	checkFindings(t, "unused-variable", `
a = 1
b = 2
c = 3
d = (a if b else c)  # only d is unused
e = 5 # @unused
# @unused
f = 7`,
		[]string{":4: Variable \"d\" is unused."},
		scopeBuild|scopeWorkspace)

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
		scopeBuild|scopeWorkspace)

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
		scopeBuild|scopeWorkspace)
}

func TestRedefinedVariable(t *testing.T) {
	checkFindings(t, "redefined-variable", `
x = "old_value"
x = "new_value"
cc_library(name = x)`,
		[]string{":2: Variable \"x\" has already been defined."},
		scopeEverywhere)

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
		scopeEverywhere)
}

func TestPackageOnTop(t *testing.T) {
	checkFindings(t, "package-on-top", `
my_macro(name = "foo")
package()`,
		[]string{":2: Package declaration should be at the top of the file, after the load() statements, but before any call to a rule or a macro. package_group() and licenses() may be called before package()."},
		scopeEverywhere)
}

func TestLoadOnTop(t *testing.T) {
	checkFindingsAndFix(t, "load-on-top", `
foo()
load(":f.bzl", "x")
x()`, `
load(":f.bzl", "x")

foo()

x()`,
		[]string{":2: Load statements should be at the top of the file."}, scopeBuild|scopeBzl)

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
		}, scopeBuild|scopeBzl)
}

func TestOutOfOrderLoad(t *testing.T) {
	checkFindingsAndFix(t, "out-of-order-load", `
# b comment
load(":b.bzl", "b")
b += 2
# a comment
load(":a.bzl", "a")
a + b`, `
# a comment
load(":a.bzl", "a")
b += 2
# b comment
load(":b.bzl", "b")
a + b`,
		[]string{":5: Load statement is out of its lexicographical order."},
		scopeBuild|scopeBzl)

	checkFindingsAndFix(t, "out-of-order-load", `
# b comment
load(":b.bzl", "b")
# c comment
load(":c.bzl", "c")
# a comment
load(":a.bzl", "a")
a + b + c`, `
# a comment
load(":a.bzl", "a")
# b comment
load(":b.bzl", "b")
# c comment
load(":c.bzl", "c")
a + b + c`,
		[]string{":6: Load statement is out of its lexicographical order."},
		scopeBuild|scopeBzl)

	checkFindingsAndFix(t, "out-of-order-load", `
load(":a.bzl", "a")
load("//a:a.bzl", "a")
load("@a//a:a.bzl", "a")
load("//b:b.bzl", "b")
load(":b.bzl", "b")
load("@b//b:b.bzl", "b")`, `
load("@a//a:a.bzl", "a")
load("@b//b:b.bzl", "b")
load("//a:a.bzl", "a")
load("//b:b.bzl", "b")
load(":a.bzl", "a")
load(":b.bzl", "b")
`,
		[]string{
			":2: Load statement is out of its lexicographical order.",
			":3: Load statement is out of its lexicographical order.",
			":6: Load statement is out of its lexicographical order.",
		}, scopeBuild|scopeBzl)

	checkFindingsAndFix(t, "out-of-order-load", `
load(":a.bzl", "a")
load(":a.bzl", "a")
`, `
load(":a.bzl", "a")
load(":a.bzl", "a")`,
		[]string{}, scopeBuild|scopeBzl)
}

func TestUnsortedDictItems(t *testing.T) {
	checkFindingsAndFix(t, "unsorted-dict-items", `
d = {
	"b": "b value",
	"a": "a value",
}`, `
d = {
	"a": "a value",
	"b": "b value",
}`,
		[]string{":3: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
d = {
	"a": "a value",
	"a": "a value",
}`,
		[]string{},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
d = {
	2: "two",
	"b": "b value",
	1: "one",
	"a": "a value",
	3: "three",
}`, `
d = {
	2: "two",
	"a": "a value",
	1: "one",
	"b": "b value",
	3: "three",
}`,
		[]string{":5: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
d = {}`,
		[]string{},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
d = {
	# b comment
	"b": "b value",
	"a": "a value",
}`, `
d = {
	"a": "a value",
	# b comment
	"b": "b value",
}`,
		[]string{":4: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
# @unsorted-dict-items
d = {
	"b": "b value",
	"a": "a value",
}`,
		[]string{},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
d = {
	"key" : {
		"b": "b value",
		"a": "a value",
	}
}`, `
d = {
	"key" : {
		"a": "a value",
		"b": "b value",
	}
}`,
		[]string{"4: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
d = {
	"deps": [],
	"srcs": ["foo.go"],
}`, `
d = {
	"srcs": ["foo.go"],
	"deps": [],
}`,
		[]string{"3: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
foo_binary = rule(
	implementation = _foo_binary_impl,
	attrs = {
		"_foocc": attr.label(
			default = Label("//depsets:foocc"),
		),
		"srcs": attr.label_list(allow_files = True),
		"deps": attr.label_list(),
	},
	outputs = {"out": "%{name}.out"},
)`, `
foo_binary = rule(
	implementation = _foo_binary_impl,
	attrs = {
		"srcs": attr.label_list(allow_files = True),
		"deps": attr.label_list(),
		"_foocc": attr.label(
			default = Label("//depsets:foocc"),
		),
	},
	outputs = {"out": "%{name}.out"},
)`,
		[]string{"7: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
# @unsorted-dict-items
d = {
	"key" : {
		"b": "b value",
		"a": "a value",
	}
}`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
d.update(
	# @unsorted-dict-items
	{
		"b": "value2",
		"a": "value1",
	},
)`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "unsorted-dict-items", `
d.update(
	{
		"b": "value2",
		"a": "value1",
	}, # @unsorted-dict-items
)`,
		[]string{},
		scopeEverywhere)
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
		scopeEverywhere)
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
		scopeEverywhere)
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
		scopeEverywhere)
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
		scopeEverywhere)
}

func TestDepsetUnion(t *testing.T) {
	checkFindings(t, "depset-union", `
d = depset([1, 2, 3])

d + foo
foo + d
d + foo + bar
foo + bar + d

d | foo
foo | d
d | foo | bar
foo | bar | d

d += foo
d |= bar
foo += d
bar |= d

d.union(aaa)
bbb.union(d)

ccc.union(ddd)
eee + fff | ggg
`,
		[]string{
			":3: Depsets should be joined using the depset constructor",
			":4: Depsets should be joined using the depset constructor",
			":5: Depsets should be joined using the depset constructor",
			":5: Depsets should be joined using the depset constructor",
			":6: Depsets should be joined using the depset constructor",
			":8: Depsets should be joined using the depset constructor",
			":9: Depsets should be joined using the depset constructor",
			":10: Depsets should be joined using the depset constructor",
			":10: Depsets should be joined using the depset constructor",
			":11: Depsets should be joined using the depset constructor",
			":13: Depsets should be joined using the depset constructor",
			":14: Depsets should be joined using the depset constructor",
			":15: Depsets should be joined using the depset constructor",
			":16: Depsets should be joined using the depset constructor",
			":18: Depsets should be joined using the depset constructor",
			":19: Depsets should be joined using the depset constructor",
		},
		scopeEverywhere)
}

func TestArgumentsOrder(t *testing.T) {
	checkFindingsAndFix(t, "args-order", `
foo(1, a = b, c + d, **e, *f)
foo(b = c, a)
foo(*d, a)
foo(**e, a)
foo(*d, b = c)
foo(**e, b = c)
foo(**e, *d)
foo(**e, *d, b = c, b2 = c2, a, a2)
foo(bar = bar(x = y, z), baz * 2)
`, `
foo(1, c + d, a = b, *f, **e)
foo(a, b = c)
foo(a, *d)
foo(a, **e)
foo(b = c, *d)
foo(b = c, **e)
foo(*d, **e)
foo(a, a2, b = c, b2 = c2, *d, **e)
foo(baz * 2, bar = bar(z, x = y))
`,
		[]string{
			":1: Function call arguments should be in the following order",
			":2: Function call arguments should be in the following order",
			":3: Function call arguments should be in the following order",
			":4: Function call arguments should be in the following order",
			":5: Function call arguments should be in the following order",
			":6: Function call arguments should be in the following order",
			":7: Function call arguments should be in the following order",
			":8: Function call arguments should be in the following order",
			":9: Function call arguments should be in the following order",
			":9: Function call arguments should be in the following order",
		},
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
