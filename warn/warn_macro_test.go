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
  native.package_name()
  native.repository_name()
  native.exports_files()
  return my_rule

def another_macro(x):
  foo()
  [native.cc_library() for i in x]
`,
		[]string{
			`5: The macro "macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:5: macro
test/package:test_file.bzl:7: my_rule
test/package:test_file.bzl:3: (RULE)

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_macro").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
			`:19: The macro "another_macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:19: another_macro
test/package:test_file.bzl:21: native.cc_library (NATIVE RULE)

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_another_macro").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
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
			`:8: The macro "bad_macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:8: bad_macro
test/package:test_file.bzl:10: alias
test/package:test_file.bzl:6: alias
test/package:test_file.bzl:3: macro
test/package:test_file.bzl:4: my_rule
test/package:test_file.bzl:1: (RULE)`,
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
			`:3: The macro "macro1" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:3: macro1
test/package:test_file.bzl:4: my_rule
test/package:test_file.bzl:1: (RULE)`,
			`:6: The macro "macro2" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:6: macro2
test/package:test_file.bzl:7: macro1
test/package:test_file.bzl:3: macro1
test/package:test_file.bzl:4: my_rule
test/package:test_file.bzl:1: (RULE)
`,
			`:9: The macro "macro3" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:9: macro3
test/package:test_file.bzl:10: macro2
test/package:test_file.bzl:6: macro2
test/package:test_file.bzl:7: macro1
test/package:test_file.bzl:3: macro1
test/package:test_file.bzl:4: my_rule
test/package:test_file.bzl:1: (RULE)
`,
			`:12: The macro "macro4" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:12: macro4
test/package:test_file.bzl:13: my_rule
test/package:test_file.bzl:1: (RULE)`,
		},
		scopeBzl)
}

func TestSymbolicMacro(t *testing.T) {
	checkFindings(t, "unnamed-macro", `
load(":foo.bzl", "foo")

my_rule = rule()

def _my_macro_implementation(name):
  my_rule()

my_symbolic_macro = macro(implementation = _my_macro_impl)

def legacy_macro():
  my_symbolic_macro()
`,
		[]string{
			`:10: The macro "legacy_macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:10: legacy_macro
test/package:test_file.bzl:11: my_symbolic_macro
test/package:test_file.bzl:8: (MACRO)
`},
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
			`:3: The macro "foo" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:3: foo
test/package:test_file.bzl:4: bar
test/package:test_file.bzl:6: bar
test/package:test_file.bzl:8: my_rule
test/package:test_file.bzl:1: (RULE)`,
			`:6: The macro "bar" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:6: bar
test/package:test_file.bzl:8: my_rule
test/package:test_file.bzl:1: (RULE)`,
		},
		scopeBzl)
}

func TestUnnamedMacroWithReader(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/subdir1/foo.bzl": `
def foo():
  native.foo_binary()

def bar():
  foo()

my_rule = rule()
`,
		"test/package/subdir2/baz.bzl": `
load(":subdir1/foo.bzl", "bar", your_rule = "my_rule")
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
load("//test/package:subdir1/foo.bzl", abc = "bar")
load(":subdir2/baz.bzl", "baz", "qux", "f")

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
			`:4: The macro "macro1" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:4: macro1
test/package:test_file.bzl:5: abc
test/package:subdir1/foo.bzl:5: bar
test/package:subdir1/foo.bzl:6: foo
test/package:subdir1/foo.bzl:2: foo
test/package:subdir1/foo.bzl:3: native.foo_binary (NATIVE RULE)

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_macro1").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
			`:7: The macro "macro2" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:7: macro2
test/package:test_file.bzl:8: baz
test/package:subdir2/baz.bzl:5: baz
test/package:subdir2/baz.bzl:7: bar
test/package:subdir1/foo.bzl:5: bar
test/package:subdir1/foo.bzl:6: foo
test/package:subdir1/foo.bzl:2: foo
test/package:subdir1/foo.bzl:3: native.foo_binary (NATIVE RULE)

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_macro2").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
			`:10: The macro "macro3" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:10: macro3
test/package:test_file.bzl:11: qux
test/package:subdir2/baz.bzl:9: qux
test/package:subdir2/baz.bzl:10: your_rule
test/package:subdir1/foo.bzl:8: (RULE)

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_macro3").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
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
		`:4: The macro "macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:4: macro
test/package:test_file.bzl:7: quux
test/package:bar.bzl:7: qux
test/package:bar.bzl:10: quuux
test/package:foo.bzl:10: qux
test/package:foo.bzl:11: native.cc_library (NATIVE RULE)`,
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
		`:5: The macro "macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:5: macro
test/package:test_file.bzl:6: bar
test/package:foo.bzl:4: foo
test/package:foo.bzl:5: some_rule
test/package:test_file.bzl:3: (RULE)`,
	}, scopeBzl)
}

func TestUnnamedMacroAliases(t *testing.T) {
	defer setUpFileReader(map[string]string{
		"test/package/foo.bzl": `
load(":bar.bzl", _bar = "bar")

bar = _bar`,
		"test/package/bar.bzl": `
my_rule = rule()

bar = my_rule`,
	})()

	checkFindings(t, "unnamed-macro", `
load(":bar.bzl", "bar")

baz = bar

def macro1():
  baz()

def macro2(name):
  baz()
`, []string{
		`:5: The macro "macro1" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:5: macro1
test/package:test_file.bzl:6: baz
test/package:test_file.bzl:3: baz
test/package:bar.bzl:4: bar
test/package:bar.bzl:2: (RULE)`,
	}, scopeBzl)
}

func TestUnnamedMacroPrivate(t *testing.T) {
	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def _not_macro(x):
  my_rule(name = x)

def macro(x):
  _not_macro(x)
`,
		[]string{
			`:6: The macro "macro" should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

test/package:test_file.bzl:6: macro
test/package:test_file.bzl:7: _not_macro
test/package:test_file.bzl:3: _not_macro
test/package:test_file.bzl:4: my_rule
test/package:test_file.bzl:1: (RULE)`,
		},
		scopeBzl)
}

func TestUnnamedMacroTypeAnnotation(t *testing.T) {
	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro(name: string):
  my_rule(name)
`,
		[]string{},
		scopeBzl)

	checkFindings(t, "unnamed-macro", `
my_rule = rule()

def macro(name: string = "default"):
  my_rule(name)
`,
		[]string{},
		scopeBzl)
}
