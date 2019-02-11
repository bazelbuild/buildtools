package warn

import "testing"

func TestMissingReturnValueWarning(t *testing.T) {
	// empty return
	checkFindings(t, "return-value", `
def foo():
  if x:
    return x
  else:
    return
`, []string{
		`:5: Some but not all execution paths of "foo" return a value.`,
	}, scopeEverywhere)

	// empty return; implicit return in the end
	checkFindings(t, "return-value", `
def bar():
  if x:
    pass
  elif y:
    return y
  else:
    for z in t:
      return
`, []string{
		`:1: Some but not all execution paths of "bar" return a value.
The function may terminate by an implicit return in the end.`,
		`:8: Some but not all execution paths of "bar" return a value.`,
	}, scopeEverywhere)

	// implicit return in the end
	checkFindings(t, "return-value", `
def foo():
  if x:
    return x
  else:
    bar()
`, []string{
		`:1: Some but not all execution paths of "foo" return a value.
The function may terminate by an implicit return in the end.`,
	}, scopeEverywhere)

	// implicit return in the end
	checkFindings(t, "return-value", `
def bar():
  if x:
    return x
  elif y:
    return y
  else:
    foo
    if z:
      return z

  if foo:
     return not foo
`, []string{
		`:1: Some but not all execution paths of "bar" return a value.
The function may terminate by an implicit return in the end.`,
	}, scopeEverywhere)

	// only returned values and fail() statements, ok
	checkFindings(t, "return-value", `
def bar():
  if x:
    return x
  elif y:
    return y
  else:
    foo
    if z:
      return z

  if foo:
     return not foo
  else:
     fail("unreachable")
`, []string{}, scopeEverywhere)

	// implicit return in the end because fail() is not a statement
	checkFindings(t, "return-value", `
def bar():
  if x:
    return x
  elif y:
    return y
  else:
    foo
    if z:
      return z

  if foo:
     return not foo
  else:
     foo() or fail("unreachable")
`, []string{
		`:1: Some but not all execution paths of "bar" return a value.
The function may terminate by an implicit return in the end.`,
	}, scopeEverywhere)

	// only empty returns, ok
	checkFindings(t, "return-value", `
def bar():
  if x:
    x()
  elif y:
    return
  else:
    foo
    if z:
      fail()

  if foo:
     return
`, []string{}, scopeEverywhere)

	// no returns, ok
	checkFindings(t, "return-value", `
def foobar():
  pass
`, []string{}, scopeEverywhere)

	// only fails, ok
	checkFindings(t, "return-value", `
def foobar():
  if foo:
    fail()
`, []string{}, scopeEverywhere)
}

func TestUnreachableStatementWarning(t *testing.T) {
	// after return
	checkFindings(t, "unreachable", `
def foo():
  return
  bar()
  baz()
`, []string{
		`:3: The statement is unreachable.`,
	}, scopeEverywhere)

	// two returns
	checkFindings(t, "unreachable", `
def foo():
  return 1
  return 2
`, []string{
		`:3: The statement is unreachable.`,
	}, scopeEverywhere)

	// after fail()
	checkFindings(t, "unreachable", `
def foo():
  fail("die")
  bar()
  baz()
`, []string{
		`:3: The statement is unreachable.`,
	}, scopeEverywhere)

	// after break and continue
	checkFindings(t, "unreachable", `
def foo():
  for x in y:
    if x:
      break
      bar()  # unreachable
    if y:
      continue
      bar()  # unreachable

def bar():
  for x in y:
    if x:
      break
    elif y:
      continue
    else:
      return x

    foo()  # unreachable
  foobar()  # potentially reachable
`, []string{
		`:5: The statement is unreachable.`,
		`:8: The statement is unreachable.`,
		`:19: The statement is unreachable.`,
	}, scopeEverywhere)

	// ok
	checkFindings(t, "unreachable", `
def foo():
  if x:
    return
  bar()
`, []string{}, scopeEverywhere)

	// ok
	checkFindings(t, "unreachable", `
def foo():
  x() or fail("maybe")
  bar()
`, []string{}, scopeEverywhere)
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

func TestWarnUnusedVariable(t *testing.T) {
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

func TestWarnUnusedLoad(t *testing.T) {
	checkFindingsAndFix(t, "load", `
load(":f.bzl", "s1", "s2")
load(":bar.bzl", "s1")
foo(name = s1)`, `
load(":f.bzl", "s1")
load(":bar.bzl", "s1")
foo(name = s1)`,
		[]string{
			":1: Loaded symbol \"s2\" is unused.",
			":2: Symbol \"s1\" has already been loaded.",
		},
		scopeEverywhere)

	checkFindingsAndFix(t, "load", `
load("foo", "a", "b", "c")
load("foo", "a", "d", "e")

z = a + b + d`, `
load("foo", "a", "b")
load("foo", "d")

z = a + b + d`,
		[]string{
			":1: Loaded symbol \"c\" is unused.",
			":2: Symbol \"a\" has already been loaded.",
			":2: Loaded symbol \"e\" is unused.",
		},
		scopeEverywhere)

	checkFindingsAndFix(t, "load", `
load("foo", "a")
a(1)
load("bar", "a")
a(2)
load("bar", a = "a")
a(3)
load("bar", a = "b")
a(4)
load("foo", "a")
a(5)
load("foo", "a")
a(6)
load("foo", a = "a")
a(7)`, `
load("foo", "a")
a(1)
load("bar", "a")
a(2)

a(3)
load("bar", a = "b")
a(4)
load("foo", "a")
a(5)

a(6)

a(7)`,
		[]string{
			":3: Symbol \"a\" has already been loaded.",
			":5: Symbol \"a\" has already been loaded.",
			":7: Symbol \"a\" has already been loaded.",
			":9: Symbol \"a\" has already been loaded.",
			":11: Symbol \"a\" has already been loaded.",
			":13: Symbol \"a\" has already been loaded.",
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

func TestUninitializedVariable(t *testing.T) {
	checkFindings(t, "uninitialized", `
def foo(x):
  if bar:
    x = 1
    y = 2

  bar = True
  print(x + y)
`,
		[]string{
			":2: Variable \"bar\" may not have been initialized.",
			":7: Variable \"y\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(x):
  for t in s:
    x = 1
    y = 2

  print(x + y)
`,
		[]string{
			":6: Variable \"y\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo():
  if bar:
    x = 1
    y = 2
  else:
    if foobar:
      x = 3
      y = 4
    else:
      x = 5

  print(x + y)
`,
		[]string{
			":12: Variable \"y\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(x):
  if bar:
    t = 1
  else:
    for t in maybe_empty:  
      pass

  print(t)
`,
		[]string{
			":8: Variable \"t\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(x):
  if bar:
    for t in [2, 3]:
      pass

  print(t)
`,
		[]string{
			":6: Variable \"t\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(x):
  print(t)  # maybe global or loaded
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(x):
  if bar:
    y = 1

  print(y)
  x, y = y, x
  print(y)
`,
		[]string{
			":5: Variable \"y\" may not have been initialized.",
			":6: Variable \"y\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo():
  if a:
    x = 1
    y = 1
    z = 1
  elif b:
    x = 2
    z = 2
    t = 2
  else:
    x = 3
    y = 3
    t = 3

  print(x + y + z + t)
`,
		[]string{
			":15: Variable \"y\" may not have been initialized.",
			":15: Variable \"z\" may not have been initialized.",
			":15: Variable \"t\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(y):
  if y < 0:
    x = -1
  elif y > 0:
    x = 1
  else:
    fail()

  print(x)
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(y):
  if y < 0:
    x = -1
  elif y > 0:
    x = 1
  else:
    if z:
      fail("z")
    else:
      fail("not z")

  print(x)
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(y):
  if y < 0:
    return
  elif y > 0:
    x = 1
  else:
    return x  # not initialized

  print(x)
`,
		[]string{
			":7: Variable \"x\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo(y):
  if y < 0:
    return
  elif y > 0:
    x = 1
  else:
    pass

  print(x)  # not initialized
`,
		[]string{
			":9: Variable \"x\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo():
  for x in y:
    print(x)
    print(x.attr)

  print(x)
  print(x.attr)
`,
		[]string{
			":6: Variable \"x\" may not have been initialized.",
			":7: Variable \"x\" may not have been initialized.",
		},
		scopeEverywhere)

	checkFindings(t, "uninitialized", `
def foo():
  for x in y:
    a = x
    print(a)

  print(a)
`,
		[]string{
			":6: Variable \"a\" may not have been initialized.",
		},
		scopeEverywhere)
}
