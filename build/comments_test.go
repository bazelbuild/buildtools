/*
Copyright 2026 Google LLC

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

func callByName(name string) func(f *File) Expr {
	return func(f *File) Expr {
		var expr Expr
		WalkInterruptable(f, func(x Expr, stk []Expr) error {
			if call, ok := x.(*CallExpr); ok {
				if ident, ok := call.X.(*Ident); ok && ident.Name == name {
					expr = call
					return &StopTraversalError{}
				}
			}
			return nil
		})
		return expr
	}
}

func assignExprByLHSName(name string) func(f *File) Expr {
	return func(f *File) Expr {
		var expr Expr
		WalkInterruptable(f, func(x Expr, stk []Expr) error {
			if assign, ok := x.(*AssignExpr); ok {
				if ident, ok := assign.LHS.(*Ident); ok && ident.Name == name {
					expr = assign
					return &StopTraversalError{}
				}
			}
			return nil
		})
		return expr
	}
}

func TestHasCommentContaining(t *testing.T) {
	var tests = []struct {
		name      string
		buildFile string
		selector  func(*File) Expr
		comment   string
		want      bool
	}{
		{
			name: "rule_call_with_comment",
			buildFile: `
# has-comment
my_rule(
    name = "my_target",
)`,
			comment:  "has-comment",
			selector: callByName("my_rule"),
			want:     true,
		},
		{
			name: "rule_call_with_trailing_comment",
			buildFile: `
my_rule(
    name = "my_target",
) # has-comment
`,
			comment:  "has-comment",
			selector: callByName("my_rule"),
			want:     true,
		},
		{
			name: "rule_call_with_inherited_comment",
			buildFile: `
# has-comment
my_rule(
    name = "my_target",
		attr = my_func()
)
`,
			comment:  "has-comment",
			selector: callByName("my_func"),
			want:     true,
		},
		{
			name: "function_call_with_comment",
			buildFile: `
my_rule(
    name = "my_target",
		# has-comment
		attr = my_func()
)
`,
			comment:  "has-comment",
			selector: callByName("my_func"),
			want:     true,
		},
		{
			name: "function_call_with_trailing_comment",
			buildFile: `
my_rule(
    name = "my_target",
		attr = my_func() # has-comment
)
`,
			comment:  "has-comment",
			selector: callByName("my_func"),
			want:     true,
		},
		{
			name: "rule_call_without_comment",
			buildFile: `
my_rule(
    name = "my_target",
)`,
			comment:  "has-comment",
			selector: callByName("my_rule"),
			want:     false,
		},
		{
			name: "sibling_call_has_comment",
			buildFile: `
my_rule(
    name = "my_target",
		attr = my_func(),
		attr2 = my_other_func() # has-comment
)`,
			comment:  "has-comment",
			selector: callByName("my_func"),
			want:     false,
		},
		{
			name: "assign_expr_with_trailing_comment",
			buildFile: `
my_var = 1 # has-comment
`,
			comment:  "has-comment",
			selector: assignExprByLHSName("my_var"),
			want:     true,
		},
		{
			name: "assign_expr_with_inherited_comment",
			buildFile: `
# has-comment
my_rule(
	name = "my_target",
	attr = my_func(func_arg = 1)
)
`,
			comment:  "has-comment",
			selector: assignExprByLHSName("func_arg"),
			want:     true,
		},
		{
			name: "assign_expr_without_comment",
			buildFile: `
my_var = 1
`,
			comment:  "has-comment",
			selector: assignExprByLHSName("my_var"),
			want:     false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bld, err := Parse("BUILD", []byte(tc.buildFile))
			if err != nil {
				t.Error(err)
			}
			expr := tc.selector(bld)
			if expr == nil {
				t.Error("selector returned nil")
			}
			got := HasCommentContaining(expr, tc.comment)
			if got != tc.want {
				t.Errorf("HasCommentContaining(%q) = %v, want %v", tc.comment, got, tc.want)
			}
		})
	}
}
