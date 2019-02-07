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
  name = "name",
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
