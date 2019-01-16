package warn

import "testing"

func TestDepsetIteration(t *testing.T) {
	checkFindingsAndFix(t, "depset-iteration", `
d = depset([1, 2, 3]) + bar

max(d + foo)
min(d)
all(d)
any(d)
sorted(d)
zip(
    d,
    a,
    b,
)
zip(
     a,
     d,
)
list(d)
tuple(d)
depset(d)
len(d)
1 in d
2 not in d

[foo(x) for x in d]

for x in d:
    pass

# Non-iteration is ok

foobar(d)
d == b

# The following iterations over a list don't trigger warnings

l = list([1, 2, 3])

max(l)
zip(l, foo)
[foo(x) for x in l]
1 in l

for x in l:
    pass
`, `
d = depset([1, 2, 3]) + bar

max((d + foo).to_list())
min(d.to_list())
all(d.to_list())
any(d.to_list())
sorted(d.to_list())
zip(
    d.to_list(),
    a,
    b,
)
zip(
    a,
    d.to_list(),
)
d.to_list()
tuple(d.to_list())
depset(d.to_list())
len(d.to_list())
1 in d.to_list()
2 not in d.to_list()

[foo(x) for x in d.to_list()]

for x in d.to_list():
    pass

# Non-iteration is ok

foobar(d)
d == b

# The following iterations over a list don't trigger warnings

l = list([1, 2, 3])

max(l)
zip(l, foo)
[foo(x) for x in l]
1 in l

for x in l:
    pass
`,
		[]string{
			":3: Depset iteration is deprecated.",
			":4: Depset iteration is deprecated.",
			":5: Depset iteration is deprecated.",
			":6: Depset iteration is deprecated.",
			":7: Depset iteration is deprecated.",
			":9: Depset iteration is deprecated.",
			":15: Depset iteration is deprecated.",
			":17: Depset iteration is deprecated.",
			":18: Depset iteration is deprecated.",
			":19: Depset iteration is deprecated.",
			":20: Depset iteration is deprecated.",
			":21: Depset iteration is deprecated.",
			":22: Depset iteration is deprecated.",
			":24: Depset iteration is deprecated.",
			":26: Depset iteration is deprecated.",
		},
		scopeEverywhere)
}

func TestDepsetUnion(t *testing.T) {
	checkFindings(t, "depset-union", `
d = depset([1, 2, 3])

d + foo
foo + d
d + foo + bar
foo + bar + d

d | foo
foo | d
d | foo | bar
foo | bar | d

d += foo
d |= bar
foo += d
bar |= d

d.union(aaa)
bbb.union(d)

ccc.union(ddd)
eee + fff | ggg
`,
		[]string{
			":3: Depsets should be joined using the depset constructor",
			":4: Depsets should be joined using the depset constructor",
			":5: Depsets should be joined using the depset constructor",
			":5: Depsets should be joined using the depset constructor",
			":6: Depsets should be joined using the depset constructor",
			":8: Depsets should be joined using the depset constructor",
			":9: Depsets should be joined using the depset constructor",
			":10: Depsets should be joined using the depset constructor",
			":10: Depsets should be joined using the depset constructor",
			":11: Depsets should be joined using the depset constructor",
			":13: Depsets should be joined using the depset constructor",
			":14: Depsets should be joined using the depset constructor",
			":15: Depsets should be joined using the depset constructor",
			":16: Depsets should be joined using the depset constructor",
			":18: Depsets should be joined using the depset constructor",
			":19: Depsets should be joined using the depset constructor",
		},
		scopeEverywhere)
}
