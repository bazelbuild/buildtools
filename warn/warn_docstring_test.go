package warn

import "testing"

func TestModuleDocstring(t *testing.T) {
	checkFindings(t, "module-docstring", ``,
		[]string{},
		scopeBzl)

	checkFindings(t, "module-docstring", `
# empty file`,
		[]string{},
		scopeBzl)

	checkFindings(t, "module-docstring", `
"""This is the module"""

load("foo", "bar")

bar()`,
		[]string{},
		scopeBzl)

	checkFindings(t, "module-docstring", `
load("foo", "bar")

"""This is the module"""

bar()`,
		[]string{":1: The file has no module docstring."},
		scopeBzl)

	checkFindings(t, "module-docstring", `
# comment

# comment
"""This is the module"""

load("foo", "bar")

bar()`,
		[]string{},
		scopeBzl)

	checkFindings(t, "module-docstring", `
# comment

load("foo", "bar")

# comment
"""This is the module"""

bar()`,
		[]string{":3: The file has no module docstring."},
		scopeBzl)
}

func TestFunctionDocstringExists(t *testing.T) {
	checkFindings(t, "function-docstring", `
def f(x):
   # short function
   return x
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x):
   """Short function with a docstring"""
   return x
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x, y):
   """Short function with a docstring

   Arguments:
     x: smth
   """
   return x + y
`,
		[]string{
			`2: Argument "y" is not documented.`,
			`4: Prefer 'Args:' to 'Arguments:' when documenting function arguments.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x):
   # long function
   x += 1
   x *= 2
   x /= 3
   x -= 4
   x %= 5
   return x
`,
		[]string{":1: The function \"f\" has no docstring."},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def _f(x):
   # long private function
   x += 1
   x *= 2
   x /= 3
   x -= 4
   x %= 5
   return x
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def _f(x):
   """Long private function
   with a docstring"""
   x += 1
   x *= 2
   x /= 3
   x -= 4
   x %= 5
   return x
`,
		[]string{
			`:2: The docstring for the function "_f" should start with a one-line summary.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def _f(x, y):
   """Long private function
   
   Args:
     x: something
     z: something
   """
   x *= 2
   x /= 3
   x -= 4
   x %= 5
   return x
`,
		[]string{
			`:2: Argument "y" is not documented.`,
			`:6: Argument "z" is documented but doesn't exist in the function signature.`,
		},
		scopeEverywhere)
}

func TestFunctionDocstringFormat(t *testing.T) {
	checkFindings(t, "function-docstring", `
def f(x):
   """This is a function.

   Args:
     x: something

   Returns:
     something
   """
   pass
   pass
   pass
   pass
   pass
   return x
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x):
   """This is a function.

   Args:
     x: something
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x):
   """This is a function.

   Args:
     x: something
   """
   pass
   pass
   pass
   pass
   pass
   return x
`,
		[]string{`2: Return value of "f" is not documented.`},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x):
   """This is a function.

   Args:
     x: something
   """
   pass
   pass
   pass
   pass
   pass
   if foo:
     return
   else:
     return x
`,
		[]string{`2: Return value of "f" is not documented.`},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x, y):
   """This is a function.

   Arguments:
     x: something
        y: something (this is in fact the description of x continued)
     z: something else

   Returns:
     None
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{
			`2: Argument "y" is not documented.`,
			`4: Prefer 'Args:' to 'Arguments:' when documenting function arguments.`,
			`7: Argument "z" is documented but doesn't exist in the function signature.`,
		}, scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x, y, z = None, *args, **kwargs):
   """This is a function.
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{
			`2: Arguments "x", "y", "z", "*args", "**kwargs" are not documented.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f(x, y, z = None, *args, **kwargs):
   """This is a function.

   Args:
    x: x
    y (deprecated, mutable): y
    z: z
    *args: the args
    **kwargs: the kwargs
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring", `
def f():
   """This is a function.
   this is the description
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{`2: The docstring for the function "f" should start with a one-line summary.`},
		scopeEverywhere)
}
