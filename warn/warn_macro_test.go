package warn

import "testing"

func TestUnnamedMacroNoReaderSameFile(t *testing.T) {
	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "foo")

my_rule = rule()

def macro(x):
  foo()
  my_rule(name = x)

def not_macro(x):
  foo()
  native.glob()
  native.existing_rule()
  native.existing_rules()
  return my_rule

def another_macro(x):
  foo()
  [native.cc_library() for i in x]
`,
		[]string{
			`5: Macro function "macro" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 7, in macro
    my_rule(...)
  File "test/package/test_file.bzl", line 3
    my_rule = rule(...)`,
			`16: Macro function "another_macro" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 18, in another_macro
    native.cc_library(...)`,
		},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1(foo, name, bar):
  my_rule()

def macro2(foo, *, name):
  my_rule()

def macro3(foo, *args, **kwargs):
  my_rule()
`,
		[]string{},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro(name):
  my_rule(name = name)

alias = macro

def bad_macro():
  for x in y:
    alias(x)
`,
		[]string{
			`8: Macro function "bad_macro" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 10, in bad_macro
    alias(...)
  File "test/package/test_file.bzl", line 6
    alias = macro
  File "test/package/test_file.bzl", line 4, in macro
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
		},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1():
  my_rule(name = x)

def macro2():
  macro1()

def macro3():
  macro2()

def macro4():
  my_rule()
`,
		[]string{
			`3: Macro function "macro1" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 4, in macro1
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
			`6: Macro function "macro2" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 7, in macro2
    macro1(...)
  File "test/package/test_file.bzl", line 4, in macro1
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
			`9: Macro function "macro3" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 10, in macro3
    macro2(...)
  File "test/package/test_file.bzl", line 7, in macro2
    macro1(...)
  File "test/package/test_file.bzl", line 4, in macro1
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
			`12: Macro function "macro4" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 13, in macro4
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
		},
		scopeBzl)
}

func TestUnnamedMacroRecursion(t *testing.T) {
	// Recursion is not allowed in Bazel, but shouldn't cause Buildifier to crash

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro():
  macro()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro():
  macro()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro1():
  macro2()

def macro2():
  macro3()

def macro3():
  macro1()
`, []string{}, scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def foo():
  bar()

def bar():
  foo()
  my_rule()
`,
		[]string{
			`3: Macro function "foo" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 4, in foo
    bar(...)
  File "test/package/test_file.bzl", line 8, in bar
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
			`6: Macro function "bar" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 8, in bar
    my_rule(...)
  File "test/package/test_file.bzl", line 1
    my_rule = rule(...)`,
		},
		scopeBzl)
}

func TestUnnamedMacroWithReader(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
def foo():
  native.foo_binary()

def bar():
  foo()

my_rule = rule()
`,
		"test/package/baz.bzl": `
load(":foo.bzl", "bar", your_rule = "my_rule")
load("//does/not:exist.bzl", "something")

def baz():
  if False:
    bar()

def qux():
  your_rule()

def f():
  something()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load("//test/package/foo.bzl", abc = "bar")
load(":baz.bzl", "baz", "qux", "f")

def macro1(surname):
  abc()

def macro2(surname):
  baz()

def macro3(surname):
  qux()

def not_macro(x):
  f()
`,
		[]string{
			`4: Macro function "macro1" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 5, in macro1
    abc(...)
  File "test/package/foo.bzl", line 6, in bar
    foo(...)
  File "test/package/foo.bzl", line 3, in foo
    native.foo_binary(...)`,
			`7: Macro function "macro2" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 8, in macro2
    baz(...)
  File "test/package/baz.bzl", line 7, in baz
    bar(...)
  File "test/package/foo.bzl", line 6, in bar
    foo(...)
  File "test/package/foo.bzl", line 3, in foo
    native.foo_binary(...)`,
			`10: Macro function "macro3" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 11, in macro3
    qux(...)
  File "test/package/baz.bzl", line 10, in qux
    your_rule(...)
  File "test/package/foo.bzl", line 8
    my_rule = rule(...)`,
		},
		scopeBzl)
}

func TestUnnamedMacroRecursionWithReader(t *testing.T) {
	// Recursion is not allowed in Bazel, but shouldn't cause Buildifier to crash

	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":bar.bzl", "bar")

def foo():
  foo()

def baz():
  bar()

def qux():
  native.cc_library()

`,
		"test/package/bar.bzl": `
load(":foo.bzl", "foo", "baz", quuux = "qux")

def bar():
  baz()

def qux():
  foo()
  baz()
  quuux()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "foo", "baz")
load(":bar.bzl", quux = "qux")

def macro():
  foo()
  baz()
  quux()
`, []string{
		`4: Macro function "macro" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 7, in macro
    quux(...)
  File "test/package/bar.bzl", line 10, in qux
    quuux(...)
  File "test/package/foo.bzl", line 11, in qux
    native.cc_library(...)`,
	}, scopeBzl)
}

func TestUnnamedMacroLoadCycle(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":test_file.bzl", some_rule = "my_rule")

def foo():
  some_rule()
`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", bar = "foo")

my_rule = rule()

def macro():
  bar()
`, []string{
		`5: Macro function "macro" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 6, in macro
    bar(...)
  File "test/package/foo.bzl", line 5, in foo
    some_rule(...)
  File "test/package/test_file.bzl", line 3
    my_rule = rule(...)`,
	}, scopeBzl)
}

func TestUnnamedMacroLoadedFiles(t *testing.T) {
	// Test that not necessary files are not loaded

	defer setUpFileReader(map[string]string{
		"a.bzl": "a = rule()",
		"b.bzl": "b = rule()",
		"c.bzl": "c = rule()",
		"d.bzl": "d = rule()",
	})()

	checkFindings(t, "unnamed-macro", `
load("//a.bzl", "a")
load("//b.bzl", "b")
load("//c.bzl", "c")
load("//d.bzl", "d")

def macro1():
  a()  # has to load a.bzl to analyze

def macro2():
  b()  # can skip b.bzl because there's a native rule
  native.cc_library()

def macro3():
  c()  # can skip c.bzl because a.bzl has already been loaded
  a()

def macro4():
  d()  # can skip d.bzl because there's a rule defined in the same file
  r()

r = rule()
`, []string{
		`6: Macro function "macro1" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 7, in macro1
    a(...)
  File "a.bzl", line 1
    a = rule(...)`,
		`9: Macro function "macro2" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 11, in macro2
    native.cc_library(...)`,
		`13: Macro function "macro3" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 15, in macro3
    a(...)
  File "a.bzl", line 1
    a = rule(...)`,
		`17: Macro function "macro4" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 19, in macro4
    r(...)
  File "test/package/test_file.bzl", line 21
    r = rule(...)`,
	}, scopeBzl)

	if len(fileReaderRequests) == 1 && fileReaderRequests[0] == "a.bzl" {
		return
	}
	t.Errorf("expected to load only a.bzl, instead loaded %d files: %v", len(fileReaderRequests), fileReaderRequests)
}

func TestUnnamedMacroAliases(t *testing.T) {
	// Test that not necessary files are not loaded

	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":bar.bzl", _bar = "bar")

bar = _bar`,
		"test/package/bar.bzl": `
my_rule = rule()

bar = my_rule`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "bar")

baz = bar

def macro1():
  baz()

def macro2(name):
  baz()
`, []string{
		`5: Macro function "macro1" doesn't accept a keyword argument "name".

Example stack trace (statically analyzed):
  File "test/package/test_file.bzl", line 6, in macro1
    baz(...)
  File "test/package/test_file.bzl", line 3
    baz = bar
  File "test/package/foo.bzl", line 4
    bar = _bar
  File "test/package/bar.bzl", line 4
    bar = my_rule
  File "test/package/bar.bzl", line 2
    my_rule = rule(...)`,
	}, scopeBzl)
}
