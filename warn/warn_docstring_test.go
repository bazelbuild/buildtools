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

func TestModuleDocstring(t *testing.T) {
	checkFindings(t, "module-docstring", ``,
		[]string{},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
# empty file`,
		[]string{},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
"""This is the module"""

load("foo", "bar")

bar()`,
		[]string{},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
load("foo", "bar")

"""This is the module"""

bar()`,
		[]string{":1: The file has no module docstring."},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
# comment

# comment
"""This is the module"""

load("foo", "bar")

bar()`,
		[]string{},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
# comment

load("foo", "bar")

# comment
"""This is the module"""

bar()`,
		[]string{":3: The file has no module docstring."},
		scopeBzl|scopeDefault)

	checkFindings(t, "module-docstring", `
def foo(bar):
  if bar:
    f()
  return g()`,
		[]string{":1: The file has no module docstring."},
		scopeBzl|scopeDefault)
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
def f(x):
   def g(x):
      # long function
      x += 1
      x *= 2
      x /= 3
      x -= 4
      x %= 5
      return x
   return g
`,
		[]string{},
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
}

func TestFunctionDocstringHeader(t *testing.T) {
	checkFindings(t, "function-docstring-header", `
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

	checkFindings(t, "function-docstring-header", `
def f():
   def g():
      """This is a function.
      this is the description
      """
      pass
      pass
      pass
      pass
      pass
   return f
`,
		[]string{`3: The docstring for the function "g" should start with a one-line summary.`},
		scopeEverywhere)

	checkFindings(t, "function-docstring-header", `
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

	checkFindings(t, "function-docstring-header", `
def f(x):
  """Long function with a docstring

	Docstring
	body
	"""
  x += 1
  x *= 2
  x /= 3
  x -= 4
  x %= 5
  return x
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-header", `
def f():
   """

   This is a function.

   This is a
   multiline description"""
   pass
   pass
   pass
   pass
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-header", `
def f():
   """\r
   Header in a CRLF formatted file.\r
\r
   This is a\r
   multiline description"""
`,
		[]string{},
		scopeEverywhere)

}

func TestFunctionDocstringArgs(t *testing.T) {
	checkFindings(t, "function-docstring-args", `
def f(x):
   """This is a function.

   Documented here:
   http://example.com

   Args:
     x: something, as described at
       http://example.com

   Returns:
     something, as described at
     https://example.com
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

	checkFindings(t, "function-docstring-args", `
def f(x):
   """This is a function.

   Args:
     x: something
   """
   passf
   pass
   pass
   pass
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(x, y):
  """Short function with a docstring

  Arguments:
    x: smth
  """
  return x + y
`,
		[]string{
			`2: Argument "y" is not documented.`,
			`4: Prefer "Args:" to "Arguments:" when documenting function arguments.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f():
  def g(x, y):
    """Short function with a docstring

    Arguments:
      x: smth
    """
    return x + y
  return g
`,
		[]string{
			`3: Argument "y" is not documented.`,
			`5: Prefer "Args:" to "Arguments:" when documenting function arguments.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f():
  def g(x, y):
    """Short function with a docstring
    """
    return x + y
  return g
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
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

	checkFindings(t, "function-docstring-args", `
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
			`4: Prefer "Args:" to "Arguments:" when documenting function arguments.`,
			`7: Argument "z" is documented but doesn't exist in the function signature.`,
		}, scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def my_function(x, y, z = None, *args, **kwargs):
   """This is a function.
   """
   pass
   pass
   pass
   pass
   pass
`,
		[]string{
			`2: Arguments "x", "y", "z", "*args", "**kwargs" are not documented.

If the documentation for the arguments exists but is not recognized by Buildifier
make sure it follows the line "Args:" which has the same indentation as the opening """,
and the argument description starts with "<argument_name>:" and indented with at least
one (preferably two) space more than "Args:", for example:

    def my_function(x):
        """Function description.

        Args:
          x: argument description, can be
            multiline with additional indentation.
        """`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(x, y, z = None, *args, **kwargs):
   """This is a function.

   Args:
    x (Map[string, int]): x
    y (deprecated, mutable): y
    z: z
    *args (List<string>): the args
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

	checkFindings(t, "function-docstring-args", `
def f(x, *, y, z = None):
   """This is a function.

   Args:
    x: x
    y: y
    z: z
   """
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(x, *, y, z = None):
   """This is a function.

   Args:
    x: x
    *: a separator
    y: y
    : argument without a name
    z: z
   """
   pass
`,
		[]string{
			`6: Argument "*" is documented but doesn't exist in the function signature.`,
			`8: Argument "" is documented but doesn't exist in the function signature.`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(x):
   """
   This is a function.

   Args:

     The function signature is extremely complicated

     x: something
   Returns:
     nothing
   """
   pass
   pass
   pass
   pass
   pass
   return None
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(foobar, *bar, **baz):
  """Some function
  
  Args:
    foobar: something
    foo: something
    bar: something
    baz: something
  """
  pass
`,
		[]string{
			`:2: Arguments "*bar", "**baz" are not documented.`,
			`:6: Argument "foo" is documented but doesn't exist in the function signature.`,
			`:7: Argument "bar" is documented but doesn't exist in the function signature. Do you mean "*bar"?`,
			`:8: Argument "baz" is documented but doesn't exist in the function signature. Do you mean "**baz"?`,
		},
		scopeEverywhere)

	checkFindings(t, "function-docstring-args", `
def f(x: int, y: str, z: bool = False, *, *bar: List[int], **baz: Mapping[str, bool]):
  """Some function
  
  Args:
    x: something
    t: something
  """
  pass
`,
		[]string{
			`:2: Arguments "y", "z", "*bar", "**baz" are not documented.`,
			`:6: Argument "t" is documented but doesn't exist in the function signature.`,
		},
		scopeEverywhere)
}

func TestFunctionDocstringReturn(t *testing.T) {
	checkFindings(t, "function-docstring-return", `
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

	checkFindings(t, "function-docstring-return", `
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

	checkFindings(t, "function-docstring-return", `
def f(x):
   """This is a function.

   Args:
     x: something
   """
   def g(y):
      return y

   pass
   pass
   pass
   pass
   pass
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-return", `
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

	checkFindings(t, "function-docstring-return", `
def f():
  def g(x):
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
  return g
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "function-docstring-return", `
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
}
