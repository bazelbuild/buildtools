package build

import "testing"

func nodeToString(e Expr) string {
	if bin, ok := e.(*BinaryExpr); ok {
		return bin.Op
	}
	if lit, ok := e.(*LiteralExpr); ok {
		return lit.Token
	}
	return "unknown"
}

func TestWalk(t *testing.T) {
	// (1 + 2) * (3 - 4)
	expr := &BinaryExpr{
		X: &BinaryExpr{
			X:  &LiteralExpr{Token: "1"},
			Op: "+",
			Y:  &LiteralExpr{Token: "2"},
		},
		Op: "*",
		Y: &BinaryExpr{
			X:  &LiteralExpr{Token: "3"},
			Op: "-",
			Y:  &LiteralExpr{Token: "4"},
		},
	}

	var prefix []string
	var postfix []string
	WalkWithPostfix(expr, func(e Expr, stack []Expr) {
		if e == nil {
			postfix = append(postfix, nodeToString(stack[len(stack)-1]))
		} else {
			prefix = append(prefix, nodeToString(e))
		}
	})

	compare(t, prefix, []string{"*", "+", "1", "2", "-", "3", "4"})
	compare(t, postfix, []string{"1", "2", "+", "3", "4", "-", "*"})
}
