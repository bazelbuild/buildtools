package warn

import "testing"

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
