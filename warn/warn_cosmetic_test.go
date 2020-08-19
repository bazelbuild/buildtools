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

func TestWarnSameOriginLoad(t *testing.T) {
	category := "same-origin-load"

	checkFindingsAndFix(t, category, `
	load(
		":f.bzl",
		"s2"
	)
	load(":t.bzl", "s3")
	load(
		":f.bzl",
		"s1"
	)`, `
	load(
		":f.bzl",
		"s1",
		"s2"
	)
	load(":t.bzl", "s3")`,
		[]string{`:7: There is already a load from ":f.bzl" on line 1. Please merge all loads from the same origin into a single one.`},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	"""Module"""

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
	"""Module"""

	load(
		":f.bzl",
		"s1",
		"s2",
		"s3"
	)`,
		[]string{`:8: There is already a load from ":f.bzl" on line 3. Please merge all loads from the same origin into a single one.`,
			`:12: There is already a load from ":f.bzl" on line 3. Please merge all loads from the same origin into a single one.`},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	load(":f.bzl", "s1")
	load(":f.bzl", "s2", "s3")
	`, `
	load(":f.bzl", "s1", "s2", "s3")
  `,
		[]string{`:2: There is already a load from ":f.bzl" on line 1. Please merge all loads from the same origin into a single one.`},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	load(":g.bzl", "s0")
	load(":f.bzl", "s1")
	load(":f.bzl",
    "s2",
    "s3")
	`, `
	load(":g.bzl", "s0")
	load(
      ":f.bzl",
      "s1",
      "s2",
      "s3",
  )`,
		[]string{`:3: There is already a load from ":f.bzl" on line 2. Please merge all loads from the same origin into a single one.`},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	load(":f.bzl", "s1")
	load(":f.bzl", "s2", "s3")
	load(":f.bzl",
    "s4")
	`, `
	load(
      ":f.bzl",
      "s1",
      "s2",
      "s3",
      "s4",
  )`,
		[]string{
			`:2: There is already a load from ":f.bzl" on line 1. Please merge all loads from the same origin into a single one.`,
			`:3: There is already a load from ":f.bzl" on line 1. Please merge all loads from the same origin into a single one.`,
		}, scopeEverywhere,
	)
}

func TestPackageOnTop(t *testing.T) {
	checkFindings(t, "package-on-top", `
my_macro(name = "foo")
package()`,
		[]string{":2: Package declaration should be at the top of the file, after the load() statements, but before any call to a rule or a macro. package_group() and licenses() may be called before package()."},
		scopeEverywhere)

	checkFindings(t, "package-on-top", `
# Some comments

"""This is a docstring"""

load(":foo.bzl", "foo")
load(":bar.bzl", baz = "bar")

package()

foo(baz)
`,
		[]string{},
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
		[]string{
			":2: Load statements should be at the top of the file.",
		}, scopeDefault|scopeBzl|scopeBuild)

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
		}, scopeDefault|scopeBzl|scopeBuild)

	checkFindingsAndFix(t, "load-on-top", `
load(":f.bzl", "x")
# after-comment

x()

load(":g.bzl", "y")
`, `
load(":f.bzl", "x")
# after-comment

load(":g.bzl", "y")

x()
`,
		[]string{
			":6: Load statements should be at the top of the file.",
		}, scopeDefault|scopeBzl|scopeBuild)
}

func TestOutOfOrderLoad(t *testing.T) {
	checkFindingsAndFix(t, "out-of-order-load", `
# b comment
load(":b.bzl", "b")
b += 2
# c comment
load(":c.bzl", "c")
load(":a.bzl", "a")
a + b + c`, `
# b comment
load(":b.bzl", "b")
b += 2
load(":a.bzl", "a")

# c comment
load(":c.bzl", "c")
a + b + c`,
		[]string{":6: Load statement is out of its lexicographical order."},
		scopeEverywhere)

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
		scopeEverywhere)

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
		}, scopeEverywhere)

	checkFindingsAndFix(t, "out-of-order-load", `
load(":a.bzl", "a")
load(":a.bzl", "a")
`, `
load(":a.bzl", "a")
load(":a.bzl", "a")`,
		[]string{}, scopeEverywhere)

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:xyz.bzl", "xyz")
load("//foo/bar:mno.bzl", "mno")
`, `
load("//foo:xyz.bzl", "xyz")
load("//foo/bar:mno.bzl", "mno")`,
		[]string{}, scopeEverywhere)

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:xyz.bzl", "xyz")
load("//foo2:mno.bzl", "mno")
`, `
load("//foo:xyz.bzl", "xyz")
load("//foo2:mno.bzl", "mno")`,
		[]string{}, scopeEverywhere)

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:b.bzl", "b")
load("//foo:a.bzl", "a")
`, `
load("//foo:a.bzl", "a")
load("//foo:b.bzl", "b")`,
		[]string{
			":2: Load statement is out of its lexicographical order.",
		}, scopeEverywhere)
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
	"srcs": ["foo.go"],
	"deps": [],
}`, `
d = {
	"deps": [],
	"srcs": ["foo.go"],
}`,
		[]string{"3: Dictionary items are out of their lexicographical order."},
		scopeEverywhere)

	checkFindingsAndFix(t, "unsorted-dict-items", `
d = select({
	"//conditions:zzz": ["myrule_b.sh"],
	"//conditions:default": ["myrule_default.sh"],
})`, `
d = select({
	"//conditions:zzz": ["myrule_b.sh"],
	"//conditions:default": ["myrule_default.sh"],
})`,
		[]string{},
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
		"deps": attr.label_list(),
		"srcs": attr.label_list(allow_files = True),
		"_foocc": attr.label(
			default = Label("//depsets:foocc"),
		),
	},
	outputs = {"out": "%{name}.out"},
)`,
		[]string{
			"7: Dictionary items are out of their lexicographical order.",
			"8: Dictionary items are out of their lexicographical order.",
		},
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

func TestSkylark(t *testing.T) {
	checkFindingsAndFix(t, "skylark-comment", `
# Skyline
foo()
# SkyLark

# Implemented in skylark
# Skylark
bar() # SKYLARK

# see https://docs.bazel.build/versions/master/skylark/lib/Label.html
Label()
`, `
# Skyline
foo()
# Starlark

# Implemented in starlark
# Starlark
bar() # STARLARK

# see https://docs.bazel.build/versions/master/skylark/lib/Label.html
Label()
`,
		[]string{
			`:3: "Skylark" is an outdated name of the language, please use "starlark" instead.`,
			`:7: "Skylark" is an outdated name of the language, please use "starlark" instead.`,
		},
		scopeEverywhere)

	checkFindingsAndFix(t, "skylark-comment", `
"""
Some docstring with skylark
""" # buildifier: disable=skylark-docstring

def f():
  """Some docstring with skylark"""
  # buildozer: disable=skylark-docstring
`, `
"""
Some docstring with skylark
""" # buildifier: disable=skylark-docstring

def f():
  """Some docstring with skylark"""
  # buildozer: disable=skylark-docstring
`,
		[]string{},
		scopeEverywhere)

	checkFindingsAndFix(t, "skylark-docstring", `
# Some file

"""
This is a docstring describing a skylark file
"""

def f():
  """SKYLARK"""

def l():
  """
  Returns https://docs.bazel.build/versions/master/skylark/lib/Label.html
  """
  return Label("skylark")
`, `
# Some file

"""
This is a docstring describing a starlark file
""" 

def f():
  """STARLARK"""

def l():
  """
  Returns https://docs.bazel.build/versions/master/skylark/lib/Label.html
  """
  return Label("skylark")
`,
		[]string{
			`:3: "Skylark" is an outdated name of the language, please use "starlark" instead.`,
			`:8: "Skylark" is an outdated name of the language, please use "starlark" instead.`,
		},
		scopeEverywhere)
}
