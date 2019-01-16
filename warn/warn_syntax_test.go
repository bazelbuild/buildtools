package warn

import "testing"

func TestArgumentsOrder(t *testing.T) {
	checkFindingsAndFix(t, "args-order", `
foo(1, a = b, c + d, **e, *f)
foo(b = c, a)
foo(*d, a)
foo(**e, a)
foo(*d, b = c)
foo(**e, b = c)
foo(**e, *d)
foo(**e, *d, b = c, b2 = c2, a, a2)
foo(bar = bar(x = y, z), baz * 2)
`, `
foo(1, c + d, a = b, *f, **e)
foo(a, b = c)
foo(a, *d)
foo(a, **e)
foo(b = c, *d)
foo(b = c, **e)
foo(*d, **e)
foo(a, a2, b = c, b2 = c2, *d, **e)
foo(baz * 2, bar = bar(z, x = y))
`,
		[]string{
			":1: Function call arguments should be in the following order",
			":2: Function call arguments should be in the following order",
			":3: Function call arguments should be in the following order",
			":4: Function call arguments should be in the following order",
			":5: Function call arguments should be in the following order",
			":6: Function call arguments should be in the following order",
			":7: Function call arguments should be in the following order",
			":8: Function call arguments should be in the following order",
			":9: Function call arguments should be in the following order",
			":9: Function call arguments should be in the following order",
		},
		scopeEverywhere)
}
