package warn

import "testing"

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

	checkFindingsAndFix(t, category, `
	load(":f.bzl", "s1")
	load(":f.bzl", "s2", "s3")
	`, `
	load(":f.bzl", "s1", "s2", "s3")
  `,
		[]string{":2: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one."},
		scopeEverywhere,
	)

	checkFindingsAndFix(t, category, `
	load(":f.bzl", "s1")
	load(":f.bzl",
    "s2",
    "s3")
	`, `
	load(
      ":f.bzl",
      "s1",
      "s2",
      "s3",
  )`,
		[]string{":2: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one."},
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
			":2: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one.",
			":3: There is already a load from \":f.bzl\". Please merge all loads from the same origin into a single one.",
		}, scopeEverywhere,
	)
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
		[]string{":2: Load statements should be at the top of the file."}, scopeBzl)

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
		}, scopeBzl)
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

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:xyz.bzl", "xyz")
load("//foo/bar:mno.bzl", "mno")
`, `
load("//foo:xyz.bzl", "xyz")
load("//foo/bar:mno.bzl", "mno")`,
		[]string{}, scopeBuild|scopeBzl)

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:xyz.bzl", "xyz")
load("//foo2:mno.bzl", "mno")
`, `
load("//foo:xyz.bzl", "xyz")
load("//foo2:mno.bzl", "mno")`,
		[]string{}, scopeBuild|scopeBzl)

	checkFindingsAndFix(t, "out-of-order-load", `
load("//foo:b.bzl", "b")
load("//foo:a.bzl", "a")
`, `
load("//foo:a.bzl", "a")
load("//foo:b.bzl", "b")`,
		[]string{":2: Load statement is out of its lexicographical order."}, scopeBuild|scopeBzl)
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
