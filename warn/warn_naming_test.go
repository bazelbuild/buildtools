package warn

import "testing"

func TestAmbiguousNames(t *testing.T) {
	checkFindings(t, "confusing-name", `
i = 0
I, x = 1, 2  # here
l = 3  # here
L = []
L[l] = 4
O = 5  # here

x, l = a, b  # here
a, b = x, l
`,
		[]string{
			":2: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":3: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":6: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":8: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
		}, scopeEverywhere)

	checkFindings(t, "confusing-name", `
def l():  # here
  if True:
    i = 1
    I = 2  # here
  else:
    l = 3  # here
    L = 4
  
  for O in Os:  # here
    pass

  x, l = a, b  # here
  a, b = x, l
`,
		[]string{
			":1: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":4: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":6: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":9: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":12: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
		}, scopeEverywhere)

	checkFindings(t, "confusing-name", `
[l for l in s]

cc_library(
  name = "name-conventions",
  tags = [I for I in tags if I],
)

def f(x):
  return [O for O in x]
`,
		[]string{
			":1: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":5: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":9: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
		}, scopeEverywhere)

	checkFindings(t, "confusing-name", `
[
  foo(l, I, O)
  for l in ls  # here
  if l < m
  for I, (x, O) in bar # here 2 times
  for L in Ls
  if I > l and O == 0
  if x == L
]
`,
		[]string{
			":3: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":5: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
			":5: Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').",
		}, scopeEverywhere)
}

func TestProviderNames(t *testing.T) {
	checkFindings(t, "name-conventions", `
foo = not_provider()
foo = provider
foo = provider(args)
fooInfo = provider()
FooInfoProvider = provider()
FooInfo = provider()
_FooInfo = provider()
_Foo_Info = provider()
_ = provider()
`,
		[]string{
			`:3: Provider name "foo" should be UpperCamelCase and should end with 'Info'.`,
			`:4: Provider name "fooInfo" should be UpperCamelCase and should end with 'Info'.`,
			`:5: Provider name "FooInfoProvider" should be UpperCamelCase and should end with 'Info'.`,
			`:8: Provider name "_Foo_Info" should be UpperCamelCase and should end with 'Info'.`,
			`:9: Provider name "_" should be UpperCamelCase and should end with 'Info'.`,
		}, scopeEverywhere)

	checkFindings(t, "name-conventions", `
def f(arg = provider()):
  foo = not_provider()
  foo = provider
  foo = provider()
  fooInfo = provider()
  FooInfoProvider = provider()
  FooInfo = provider()
  _FooInfo = provider()
  _Foo_Info = provider()
  _ = provider()
`,
		[]string{
			`:4: Provider name "foo" should be UpperCamelCase and should end with 'Info'.`,
			`:5: Provider name "fooInfo" should be UpperCamelCase and should end with 'Info'.`,
			`:6: Provider name "FooInfoProvider" should be UpperCamelCase and should end with 'Info'.`,
			`:9: Provider name "_Foo_Info" should be UpperCamelCase and should end with 'Info'.`,
			`:10: Provider name "_" should be UpperCamelCase and should end with 'Info'.`,
		}, scopeEverywhere)

}

func TestVariableNames(t *testing.T) {
	checkFindings(t, "name-conventions", `
_ = 0
foo = 1
FOO = 2
Foo = 3
foo, Bar, BazInfo = 4, 5, 6
foo, (BAR, _) = 7, (8, 9) 
`,
		[]string{
			`:4: Variable name "Foo" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
			`:5: Variable name "Bar" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
			`:5: Variable name "BazInfo" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
		}, scopeEverywhere)

	checkFindings(t, "name-conventions", `
def f(x, _, Arg = None):
  _ = 0
  foo = 1
  FOO = 2
  Foo = 3
  foo, Bar, BazInfo = 4, 5, 6
  foo, (BAR, _) = 7, (8, 9) 
`,
		[]string{
			`:5: Variable name "Foo" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
			`:6: Variable name "Bar" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
			`:6: Variable name "BazInfo" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`,
		}, scopeEverywhere)

	checkFindings(t, "name-conventions", `
foo(
  Bar = 1,
  ___ = 2,
  BazInfo = 3,
	baz = 4,
  FOO = 5,
)
`,
		[]string{}, scopeEverywhere)
}
