/*
Copyright 2016 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
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

var simpleCall *CallExpr = &CallExpr{
	X: &Ident{
		Name: "java_library",
	},
	List: []Expr{
		&BinaryExpr{
			X: &Ident{
				Name: "name",
			},
			Op: "=",
			Y: &StringExpr{
				Value: "x",
			},
		},
	},
}

var simpleRule *Rule = &Rule{simpleCall, ""}

var structCall *CallExpr = &CallExpr{
	X: &DotExpr{
		X: &DotExpr{
			X: &Ident{
				Name: "foo",
			},
			Name: "bar",
		},
		Name: "baz",
	},
	List: []Expr{
		&BinaryExpr{
			X: &Ident{
				Name: "name",
			},
			Op: "=",
			Y: &StringExpr{
				Value: "x",
			},
		},
	},
}

var structRule *Rule = &Rule{structCall, ""}

func TestKind(t *testing.T) {
	if simpleRule.Kind() != "java_library" {
		t.Errorf(`simpleRule.Kind() = %v, want "java_library"`, simpleRule.Kind())
	}
	if structRule.Kind() != "foo.bar.baz" {
		t.Errorf(`structRule.Kind() = %v, want "foo.bar.baz"`, structRule.Kind())
	}
}

func TestSetKind(t *testing.T) {
	rule := &Rule{
		&CallExpr{
			X: &Ident{
				Name: "java_library",
			},
			List: []Expr{
				&BinaryExpr{
					X: &Ident{
						Name: "name",
					},
					Op: "=",
					Y: &StringExpr{
						Value: "x",
					},
				},
			},
		},
		"",
	}

	rule.SetKind("java_binary")
	compare(t, rule.Call.X, &Ident{Name: "java_binary"})

	rule.SetKind("foo.bar.baz")
	compare(t, rule.Call.X, &DotExpr{
		X: &DotExpr{
			X: &Ident{
				Name: "foo",
			},
			Name: "bar",
		},
		Name: "baz",
	})
}

func TestRules(t *testing.T) {
	f := &File{
		Stmt: []Expr{
			simpleCall,
			structCall,
		},
	}

	compare(t, f.Rules(""), []*Rule{simpleRule, structRule})
	compare(t, f.Rules("java_binary"), []*Rule(nil))
	compare(t, f.Rules("java_library"), []*Rule{simpleRule})
	compare(t, f.Rules("foo.bar.baz"), []*Rule{structRule})
}

func TestRulesNested(t *testing.T) {
	f := &File{
		Stmt: []Expr{
			&ListExpr{
				List: []Expr{
					simpleCall,
					structCall,
				},
			},
		},
	}

	compare(t, f.Rules(""), []*Rule{simpleRule, structRule})
	compare(t, f.Rules("java_binary"), []*Rule(nil))
	compare(t, f.Rules("java_library"), []*Rule{simpleRule})
	compare(t, f.Rules("foo.bar.baz"), []*Rule{structRule})
}

func TestRulesDoubleNested(t *testing.T) {
	var doubleNested *CallExpr = &CallExpr{
		X: &Ident{
			Name: "java_library",
		},
		List: []Expr{
			&BinaryExpr{
				X: &Ident{
					Name: "name",
				},
				Op: "=",
				Y: &CallExpr{
					X:    &Ident{Name: "varref"},
					List: []Expr{&StringExpr{Value: "x"}},
				},
			},
		},
	}
	var doubleNestedRule *Rule = &Rule{doubleNested, ""}

	f := &File{
		Stmt: []Expr{
			&ListExpr{
				List: []Expr{
					doubleNested,
				},
			},
		},
	}

	compare(t, f.Rules(""), []*Rule{doubleNestedRule})
	compare(t, f.Rules("java_library"), []*Rule{doubleNestedRule})
}

func TestImplicitName(t *testing.T) {
	tests := []struct {
		path        string
		input       string
		want        string
		description string
	}{
		{"foo/BUILD", `rule()`, "foo", `Use an implicit name for one rule.`},
		{"foo/BUILD", `rule(name="a")
rule(name="b")
rule()`, "foo", `Use an implicit name for the one unnamed rule`},
		{"foo/BUILD", `rule()
rule()
rule()`, "", `No implicit name for multiple unnamed rules`},
		{"foo/BUILD", `load(":foo.bzl", "bar")
rule(name="a")
rule(name="b")`, "", `No implicit name for load`},
		{"foo/BUILD", `load(":foo.bzl", "bar")
rule()`, "foo", `Use an implicit name for one unnamed rule with load`},
		{"BUILD", `rule()`, "", `No implicit name for root package`},
	}

	for _, tst := range tests {
		file, err := Parse(tst.path, []byte(tst.input))
		if err != nil {
			t.Error(tst.description, err)
			continue
		}

		if got := file.implicitRuleName(); got != tst.want {
			t.Errorf("TestImplicitName(%s): got %s, want %s. %s", tst.description, got, tst.want)
		}
	}
}
