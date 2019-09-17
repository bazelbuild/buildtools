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
			`:3: Invalid escape sequence \f at position 1.`,
			`:4: Invalid escape sequences:
    \f at position 2
    \b at position 9
`,
			`:6: Invalid escape sequences:
    \a at position 9
    \b at position 11
    \c at position 13
    \d at position 15
    \e at position 17
    \f at position 19
    \g at position 21
    \h at position 23
    \i at position 25
    \j at position 27
    \k at position 29
    \l at position 31
    \m at position 33
    \o at position 37
    \p at position 39
    \q at position 41
    \s at position 45
    \u at position 49
    \v at position 51
    \w at position 53
    \y at position 59
    \z at position 61
    \8 at position 79
    \9 at position 81
`,
		},
		scopeEverywhere)

	checkFindings(t, "string-escape", `
r'foo'
r'\\foo\\"bar"\\'
r"\foo"
r'"\foo"\\\bar'

r'''
"asdf"
\a\b\c\d\e\f\g\h\i\j\k\l\m\n\o\p\q\r\s\t\u\v\w\x43\y\z\0\1\2\3\4\5\6\7\8\9
'''
`,
		[]string{},
		scopeEverywhere)
}
