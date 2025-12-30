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

package build

import "testing"

func nodeToString(e Expr) string {
	if bin, ok := e.(*BinaryExpr); ok {
		return bin.Op
	}
	if assign, ok := e.(*AssignExpr); ok {
		return assign.Op
	}
	if lit, ok := e.(*LiteralExpr); ok {
		return lit.Token
	}
	return "unknown"
}

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

func TestWalk(t *testing.T) {
	var prefix []string
	Walk(binaryExprExample, func(e Expr, stk []Expr) {
		prefix = append(prefix, nodeToString(e))
	})
	compare(t, prefix, []string{"*", "+", "1", "2", "-", "3", "4"})
}

func TestWalkOnce(t *testing.T) {
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

func TestEdit(t *testing.T) {
	expr, _ := Parse("test", []byte("1 + 2"))
	compare(t, FormatString(expr), "1 + 2\n")
	Edit(expr, func(e Expr, stk []Expr) Expr {
		// Check if there are already parens
		if len(stk) > 0 {
			if _, ok := stk[len(stk)-1].(*ParenExpr); ok {
				return nil
			}
		}
		// Add parens around literal
		if lit, ok := e.(*LiteralExpr); ok {
			lit.Start = Position{} // workaround to avoid multiline formatting
			return &ParenExpr{X: e}
		}
		return nil
	})
	compare(t, FormatString(expr), "(1) + (2)\n")
}

func TestRemoveParens(t *testing.T) {
	expr, _ := Parse("test", []byte("((((1))) + 2) + (3 + 4) * 5"))
	compare(t, FormatString(expr), "((((1))) + 2) + (3 + 4) * 5\n")
	// Remove all ParenExpr
	Edit(expr, func(e Expr, stk []Expr) Expr {
		for {
			if p, ok := e.(*ParenExpr); ok {
				e = p.X
			} else {
				return e
			}
		}
	})
	// Parens are inserted in the output due to different precedence of operators.
	compare(t, FormatString(expr), "1 + 2 + (3 + 4) * 5\n")
}
