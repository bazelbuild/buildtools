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

func TestWalkOnce(t *testing.T) {
	// (1 + 2) * (3 - 4)
	var binaryExprExample Expr = &BinaryExpr{
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

	var walk func(e *Expr)
	walk = func(e *Expr) {
		prefix = append(prefix, nodeToString(*e))
		WalkOnce(*e, walk)
		postfix = append(postfix, nodeToString(*e))
	}

	walk(&binaryExprExample)

	compare(t, prefix, []string{"*", "+", "1", "2", "-", "3", "4"})
	compare(t, postfix, []string{"1", "2", "+", "3", "4", "-", "*"})
}
