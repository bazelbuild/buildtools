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

func TestIntegerDivision(t *testing.T) {
	checkFindingsAndFix(t, "integer-division", `
a = 1
b = int(2.3)
c = 1.0
d = float(2)

e = a / b
f = a / c
g = c / a
h = c / d

a /= b
a /= c
c /= a
c /= d
`, `
a = 1
b = int(2.3)
c = 1.0
d = float(2)

e = a // b
f = a / c
g = c / a
h = c / d

a //= b
a /= c
c /= a
c /= d
`,
		[]string{
			":6: The \"/\" operator for integer division is deprecated in favor of \"//\".",
			":11: The \"/=\" operator for integer division is deprecated in favor of \"//=\".",
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

func TestListAppend(t *testing.T) {
	checkFindingsAndFix(t, "list-append", `
x = []
x += y
x += [1]
x += [2, 3]
x += [4 for y in z]
x += 5
x += [foo(
    bar,
    baz,
)]
`, `
x = []
x += y
x.append(1)
x += [2, 3]
x += [4 for y in z]
x += 5
x.append(foo(
    bar,
    baz,
))
`,
		[]string{
			`:3: Prefer using ".append()" to adding a single element list`,
			`:7: Prefer using ".append()" to adding a single element list`,
		},
		scopeEverywhere)
}

func TestDictMethodNamedArg(t *testing.T) {
	checkFindings(t, "dict-method-named-arg", `
d = dict()
d.get("a", "b")
[].get("a", default = "b")

d.get("a", default = "b") # warning
d.pop("a", default = "b") # warning
{}.setdefault("a", default = "b") # warning
`,
		[]string{
			`:5: Named argument "default" not allowed`,
			`:6: Named argument "default" not allowed`,
			`:7: Named argument "default" not allowed`,
		},
		scopeEverywhere)
}
