/*
Copyright 2016 Google LLC

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
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/testutils"
)

func TestParse(t *testing.T) {
	for i, tt := range parseTests {
		name := "test"
		if tt.out != nil {
			name = tt.out.Path
		}
		p, err := Parse(name, []byte(tt.in))
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if tt.out != nil {
			compare(t, p, tt.out)
		}
	}
}

func TestParseTestdata(t *testing.T) {
	// Test that files in the testdata directory can all be parsed.
	// For this test we don't bother checking what the tree looks like.
	// The printing tests will exercise that information.
	testdata := os.Getenv("TEST_SRCDIR") + "/" + os.Getenv("TEST_WORKSPACE") + "/build/testdata"
	outs, err := filepath.Glob(testdata + "/*")
	if err != nil {
		t.Fatal(err)
	}
	if len(outs) == 0 {
		t.Fatal("Data set is empty:", testdata)
	}
	for _, out := range outs {
		if strings.HasSuffix(out, ".error") {
			// Incorrect starlark file, skip
			continue
		}

		data, err := os.ReadFile(out)
		if err != nil {
			t.Error(err)
			continue
		}

		_, err = Parse(filepath.Base(out), data)
		if err != nil {
			t.Error(err)
		}
	}
}

// toJSON returns human-readable json for the given syntax tree.
// It is used as input to diff for comparing the actual syntax tree with the expected one.
func toJSON(v interface{}) string {
	s, _ := json.MarshalIndent(v, "", "\t")
	s = append(s, '\n')
	return string(s)
}

// Compare expected and actual values, failing and outputting a diff of the two values if they are not deeply equal
func compare(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		testutils.Tdiff(t, []byte(toJSON(expected)), []byte(toJSON(actual)))
	}
}

// Small tests checking that the parser returns exactly the right syntax tree.
// If out is nil, we only check that the parser accepts the file.
var parseTests = []struct {
	in  string
	out *File
}{
	{
		in: `go_binary(name = "x"
)
`,
		out: &File{
			Path: "BUILD",
			Type: TypeBuild,
			Stmt: []Expr{
				&CallExpr{
					X: &Ident{
						NamePos: Position{1, 1, 0},
						Name:    "go_binary",
					},
					ListStart: Position{1, 10, 9},
					List: []Expr{
						&AssignExpr{
							LHS: &Ident{
								NamePos: Position{1, 11, 10},
								Name:    "name",
							},
							OpPos: Position{1, 16, 15},
							Op:    "=",
							RHS: &StringExpr{
								Start: Position{1, 18, 17},
								Value: "x",
								End:   Position{1, 21, 20},
								Token: `"x"`,
							},
						},
					},
					End:            End{Pos: Position{2, 1, 21}},
					ForceMultiLine: true,
				},
			},
		},
	},
	{
		in: `foo.bar.baz(name = "x")`,
		out: &File{
			Path: "test",
			Type: TypeDefault,
			Stmt: []Expr{
				&CallExpr{
					X: &DotExpr{
						X: &DotExpr{
							X: &Ident{
								NamePos: Position{1, 1, 0},
								Name:    "foo",
							},
							Dot:     Position{1, 4, 3},
							NamePos: Position{1, 5, 4},
							Name:    "bar",
						},
						Dot:     Position{1, 8, 7},
						NamePos: Position{1, 9, 8},
						Name:    "baz",
					},
					ListStart: Position{1, 12, 11},
					List: []Expr{
						&AssignExpr{
							LHS: &Ident{
								NamePos: Position{1, 13, 12},
								Name:    "name",
							},
							OpPos: Position{1, 18, 17},
							Op:    "=",
							RHS: &StringExpr{
								Start: Position{1, 20, 19},
								Value: "x",
								End:   Position{1, 23, 22},
								Token: `"x"`,
							},
						},
					},
					End: End{Pos: Position{1, 23, 22}},
				},
			},
		},
	},
	{
		in: `package(default_visibility = ["//visibility:legacy_public"])
`,
	},
	{
		in: `__unused__ = [ foo_binary(
                   name = "signed_release_%sdpi" % dpi,
                   srcs = [
                       ":aps_release_%s" % dpi,  # all of Maps, obfuscated, w/o NLP
                       ":qlp_release_%s" % dpi,  # the NLP
                       ":check_binmode_release",
                       ":check_remote_strings_release",
                   ],
                   debug_key = "//foo:bar.baz",
                   resources = ":R_src_release_%sdpi" % dpi)
    for dpi in dpis ]
`,
	},
	{
		in: `load(":foo.bzl", "foo", """bar""", baz="foo", foo="""baz""")
`,
		out: &File{
			Path: "BUILD",
			Type: TypeBuild,
			Stmt: []Expr{
				&LoadStmt{
					Load: Position{1, 1, 0},
					Module: &StringExpr{
						Value: ":foo.bzl",
						Token: "\":foo.bzl\"",
						Start: Position{1, 6, 5},
						End:   Position{1, 16, 15},
					},
					From: []*Ident{
						{
							Name:    "foo",
							NamePos: Position{1, 19, 18},
						},
						{
							Name:    "bar",
							NamePos: Position{1, 28, 27},
						},
						{
							Name:    "foo",
							NamePos: Position{1, 41, 40},
						},
						{
							Name:    "baz",
							NamePos: Position{1, 54, 53},
						},
					},
					To: []*Ident{
						{
							Name:    "foo",
							NamePos: Position{1, 19, 18},
						},
						{
							Name:    "bar",
							NamePos: Position{1, 28, 27},
						},
						{
							Name:    "baz",
							NamePos: Position{1, 36, 35},
						},
						{
							Name:    "foo",
							NamePos: Position{1, 47, 46},
						},
					},
					Rparen:       End{Pos: Position{1, 60, 59}},
					ForceCompact: true,
				},
			},
		},
	},
}
