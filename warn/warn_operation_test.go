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

func TestStringEscape(t *testing.T) {
	checkFindingsAndFix(t, "string-escape", `
'foo'
'\\foo\\"bar"\\'
"\foo"
'"\foo"\\\bar'

'''
"asdf"
\a\b\c\d\e\f\g\h\i\j\k\l\m\n\o\p\q\r\s\t\u\v\w\x43\y\z\0\1\2\3\4\5\6\7\8\9
'''
`, `
"foo"
'\\foo\\"bar"\\'
"\\foo"
'"\\foo"\\\\bar'

'''
"asdf"
\\a\\b\\c\\d\\e\\f\\g\\h\\i\\j\\k\\l\\m\n\\o\\p\\q\r\\s\t\\u\\v\\w\x43\\y\\z\0\1\2\3\4\5\6\7\\8\\9
'''
`,
		[]string{
			":3: Invalid quote sequences at position 1.",
			":4: Invalid quote sequences at positions 2, 9.",
			":6: Invalid quote sequences at positions 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 37, 39, 41, 45, 49, 51, 53, 59, 61, 79, 81.",
		},
		scopeEverywhere)
}
