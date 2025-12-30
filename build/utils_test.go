/*
Copyright 2025 Google LLC

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

import (
	"testing"
)

func TestIsMultiline(t *testing.T) {
	var tests = []struct {
		name            string
		buildFile       string
		exprToCheck     func(Expr) bool
		wantIsMultiline bool
	}{
		{
			name: "single_line_call",
			exprToCheck: func(expr Expr) bool {
				if call, ok := expr.(*CallExpr); ok {
					if calledIdent, ok := call.X.(*Ident); ok && calledIdent.Name == "foo" {
						return true
					}
				}
				return false
			},
			buildFile: `
	foo(name = "bar")`,
			wantIsMultiline: false,
		},
		{
			name: "multiline_call",
			exprToCheck: func(expr Expr) bool {
				if call, ok := expr.(*CallExpr); ok {
					if calledIdent, ok := call.X.(*Ident); ok && calledIdent.Name == "foo" {
						return true
					}
				}
				return false
			},
			buildFile: `
	foo(
		name = "bar",
	)`,
			wantIsMultiline: true,
		},
		{
			name: "single_line_attribute",
			exprToCheck: func(expr Expr) bool {
				if assign, ok := expr.(*AssignExpr); ok {
					if calledIdent, ok := assign.LHS.(*Ident); ok && calledIdent.Name == "foo_attr" {
						return true
					}
				}
				return false
			},
			buildFile: `
	foo(
		name = "bar",
		foo_attr = "on_a_single_line",
	)`,
			wantIsMultiline: false,
		},
		{
			name: "multi_line_attribute",
			exprToCheck: func(expr Expr) bool {
				if assign, ok := expr.(*AssignExpr); ok {
					if calledIdent, ok := assign.LHS.(*Ident); ok && calledIdent.Name == "foo_attr" {
						return true
					}
				}
				return false
			},
			buildFile: `
	foo(
		name = "bar",
		foo_attr = [
		    "attribute",
			"which spans",
			"multiple lines",
		],
	)`,
			wantIsMultiline: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bld, err := Parse("BUILD", []byte(tc.buildFile))
			if err != nil {
				t.Fatal(err)
			}
			var testedExpr *Expr
			Walk(bld, func(x Expr, _ []Expr) {
				if tc.exprToCheck(x) {
					testedExpr = &x
				}
			})
			if testedExpr == nil {
				t.Fatal("Unable to find expression to test")
			}

			got := IsMultiLine(*testedExpr)
			if got != tc.wantIsMultiline {
				t.Fatalf("IsMultiline returned incorrect value, got: %t, expected: %t", got, tc.wantIsMultiline)
			}
		})
	}
}
