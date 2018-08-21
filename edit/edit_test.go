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
package edit

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
)

var parseLabelTests = []struct {
	in   string
	repo string
	pkg  string
	rule string
}{
	{"//devtools/buildozer:rule", "", "devtools/buildozer", "rule"},
	{"devtools/buildozer:rule", "", "devtools/buildozer", "rule"},
	{"//devtools/buildozer", "", "devtools/buildozer", "buildozer"},
	{"//base", "", "base", "base"},
	{"//base:", "", "base", "base"},
	{"@r//devtools/buildozer:rule", "r", "devtools/buildozer", "rule"},
	{"@r//devtools/buildozer", "r", "devtools/buildozer", "buildozer"},
	{"@r//base", "r", "base", "base"},
	{"@r//base:", "r", "base", "base"},
	{"@foo", "foo", "", "foo"},
	{":label", "", "", "label"},
	{"label", "", "", "label"},
	{"/abs/path/to/WORKSPACE:rule", "", "/abs/path/to/WORKSPACE", "rule"},
}

func TestParseLabel(t *testing.T) {
	for i, tt := range parseLabelTests {
		repo, pkg, rule := ParseLabel(tt.in)
		if repo != tt.repo || pkg != tt.pkg || rule != tt.rule {
			t.Errorf("%d. ParseLabel(%q) => (%q, %q, %q), want (%q, %q, %q)",
				i, tt.in, repo, pkg, rule, tt.repo, tt.pkg, tt.rule)
		}
	}
}

var shortenLabelTests = []struct {
	in     string
	pkg    string
	result string
}{
	{"//devtools/buildozer:rule", "devtools/buildozer", ":rule"},
	{"@//devtools/buildozer:rule", "devtools/buildozer", ":rule"},
	{"//devtools/buildozer:rule", "devtools", "//devtools/buildozer:rule"},
	{"//base:rule", "devtools", "//base:rule"},
	{"//base:base", "devtools", "//base"},
	{"//base", "base", ":base"},
	{"//devtools/buildozer:buildozer", "", "//devtools/buildozer"},
	{"@r//devtools/buildozer:buildozer", "devtools/buildozer", "@r//devtools/buildozer"},
	{"@r//devtools/buildozer", "devtools/buildozer", "@r//devtools/buildozer"},
	{"@r//devtools", "devtools", "@r//devtools"},
	{"@r:rule", "", "@r:rule"},
	{"@r", "", "@r"},
	{"@foo//:foo", "", "@foo"},
	{"@foo//devtools:foo", "", "@foo//devtools:foo"},
	{"@foo//devtools:foo", "devtools", "@foo//devtools:foo"},
	{"@foo//foo:foo", "", "@foo//foo"},
	{":local", "", ":local"},
	{"something else", "", "something else"},
	{"/path/to/file", "path/to", "/path/to/file"},
}

func TestShortenLabel(t *testing.T) {
	for i, tt := range shortenLabelTests {
		result := ShortenLabel(tt.in, tt.pkg)
		if result != tt.result {
			t.Errorf("%d. ShortenLabel(%q, %q) => %q, want %q",
				i, tt.in, tt.pkg, result, tt.result)
		}
	}
}

var labelsEqualTests = []struct {
	label1   string
	label2   string
	pkg      string
	expected bool
}{
	{"//devtools/buildozer:rule", "rule", "devtools/buildozer", true},
	{"//devtools/buildozer:rule", "rule:jar", "devtools", false},
}

func TestLabelsEqual(t *testing.T) {
	for i, tt := range labelsEqualTests {
		if got := LabelsEqual(tt.label1, tt.label2, tt.pkg); got != tt.expected {
			t.Errorf("%d. LabelsEqual(%q, %q, %q) => %v, want %v",
				i, tt.label1, tt.label2, tt.pkg, got, tt.expected)
		}
	}
}

var splitOnSpacesTests = []struct {
	in  string
	out []string
}{
	{"a", []string{"a"}},
	{"  abc def ", []string{"abc", "def"}},
	{`  abc\ def `, []string{"abc def"}},
}

func TestSplitOnSpaces(t *testing.T) {
	for i, tt := range splitOnSpacesTests {
		result := SplitOnSpaces(tt.in)
		if !reflect.DeepEqual(result, tt.out) {
			t.Errorf("%d. SplitOnSpaces(%q) => %q, want %q",
				i, tt.in, result, tt.out)
		}
	}
}

func TestInsertLoad(t *testing.T) {
	tests := []struct{ input, expected string }{
		{``, `load("location", "symbol")`},
		{`load("location", "symbol")`, `load("location", "symbol")`},
		{`load("location", "other", "symbol")`, `load("location", "other", "symbol")`},
		{`load("location", "other")`, `load("location", "other", "symbol")`},
		{
			`load("other loc", "symbol")`,
			`load("location", "symbol")
load("other loc", "symbol")`,
		},
	}

	for _, tst := range tests {
		bld, err := build.Parse("BUILD", []byte(tst.input))
		if err != nil {
			t.Error(err)
			continue
		}
		bld.Stmt = InsertLoad(bld.Stmt, "location", []string{"symbol"}, []string{"symbol"})
		got := strings.TrimSpace(string(build.Format(bld)))
		if got != tst.expected {
			t.Errorf("maybeInsertLoad(%s): got %s, expected %s", tst.input, got, tst.expected)
		}
	}
}

func TestAddValueToListAttribute(t *testing.T) {
	tests := []struct{ input, expected string }{
		{`rule(name="rule")`, `rule(name="rule", attr=["foo"])`},
		{`rule(name="rule", attr=["foo"])`, `rule(name="rule", attr=["foo"])`},
		{`rule(name="rule", attr=IDENT)`, `rule(name="rule", attr=IDENT+["foo"])`},
		{`rule(name="rule", attr=["foo"] + IDENT)`, `rule(name="rule", attr=["foo"] + IDENT)`},
		{`rule(name="rule", attr=["bar"] + IDENT)`, `rule(name="rule", attr=["bar", "foo"] + IDENT)`},
		{`rule(name="rule", attr=IDENT + ["foo"])`, `rule(name="rule", attr=IDENT + ["foo"])`},
		{`rule(name="rule", attr=IDENT + ["bar"])`, `rule(name="rule", attr=IDENT + ["bar", "foo"])`},
	}

	for _, tst := range tests {
		bld, err := build.Parse("BUILD", []byte(tst.input))
		if err != nil {
			t.Error(err)
			continue
		}
		rule := bld.RuleAt(1)
		AddValueToListAttribute(rule, "attr", "", &build.StringExpr{Value: "foo"}, nil)
		got := strings.TrimSpace(string(build.Format(bld)))

		wantBld, err := build.Parse("BUILD", []byte(tst.expected))
		if err != nil {
			t.Error(err)
			continue
		}
		want := strings.TrimSpace(string(build.Format(wantBld)))
		if got != want {
			t.Errorf("AddValueToListAttribute(%s): got %s, expected %s", tst.input, got, want)
		}
	}
}

func TestListSubstitute(t *testing.T) {
	tests := []struct {
		desc, input, oldPattern, newTemplate, want string
	}{
		{
			desc:        "no_match",
			input:       `["abc"]`,
			oldPattern:  `!!`,
			newTemplate: `xx`,
			want:        `["abc"]`,
		}, {
			desc:        "full_match",
			input:       `["abc"]`,
			oldPattern:  `.*`,
			newTemplate: `xx`,
			want:        `["xx"]`,
		}, {
			desc:        "partial_match",
			input:       `["abcde"]`,
			oldPattern:  `bcd`,
			newTemplate: `xyz`,
			want:        `["axyze"]`,
		}, {
			desc:        "number_group",
			input:       `["abcde"]`,
			oldPattern:  `a(bcd)`,
			newTemplate: `$1 $1`,
			want:        `["bcd bcde"]`,
		}, {
			desc:        "name_group",
			input:       `["abcde"]`,
			oldPattern:  `a(?P<x>bcd)`,
			newTemplate: `$x $x`,
			want:        `["bcd bcde"]`,
		},
	}

	for _, tst := range tests {
		t.Run(tst.desc, func(t *testing.T) {
			f, err := build.ParseBuild("BUILD", []byte(tst.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			lst := f.Stmt[0]
			oldRegexp, err := regexp.Compile(tst.oldPattern)
			if err != nil {
				t.Fatalf("error compiling regexp %q: %v", tst.oldPattern, err)
			}
			ListSubstitute(lst, oldRegexp, tst.newTemplate)
			if got := build.FormatString(lst); got != tst.want {
				t.Errorf("ListSubstitute(%q, %q, %q) = %q ; want %q", tst.input, tst.oldPattern, tst.newTemplate, got, tst.want)
			}
		})
	}
}

func compareKeyValue(a, b build.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aKeyVal := a.(*build.KeyValueExpr)
	bKeyVal := b.(*build.KeyValueExpr)
	return aKeyVal.Key.(*build.StringExpr).Value == bKeyVal.Key.(*build.StringExpr).Value &&
		aKeyVal.Value.(*build.StringExpr).Value == bKeyVal.Value.(*build.StringExpr).Value
}

func TestDictionaryDelete(t *testing.T) {
	tests := []struct {
		input, expected string
		expectedReturn  build.Expr
	}{
		{
			`rule(attr = {"deletekey": "value"})`,
			`rule(attr = {})`,
			&build.KeyValueExpr{
				Key:   &build.StringExpr{Value: "deletekey"},
				Value: &build.StringExpr{Value: "value"},
			},
		}, {
			`rule(attr = {"nodeletekey": "value", "deletekey": "value"})`,
			`rule(attr = {"nodeletekey": "value"})`,
			&build.KeyValueExpr{
				Key:   &build.StringExpr{Value: "deletekey"},
				Value: &build.StringExpr{Value: "value"},
			},
		}, {
			`rule(attr = {"nodeletekey": "value"})`,
			`rule(attr = {"nodeletekey": "value"})`,
			nil,
		},
	}

	for _, tst := range tests {
		bld, err := build.ParseBuild("BUILD", []byte(tst.input))
		if err != nil {
			t.Error(err)
			continue
		}
		rule := bld.RuleAt(1)
		dict := rule.Call.List[0].(*build.BinaryExpr).Y.(*build.DictExpr)
		returnVal := DictionaryDelete(dict, "deletekey")
		got := strings.TrimSpace(string(build.Format(bld)))
		wantBld, err := build.Parse("BUILD", []byte(tst.expected))
		if err != nil {
			t.Error(err)
			continue
		}
		want := strings.TrimSpace(string(build.Format(wantBld)))
		if got != want {
			t.Errorf("TestDictionaryDelete(%s): got %s, expected %s", tst.input, got, want)
		}
		if !compareKeyValue(returnVal, tst.expectedReturn) {
			t.Errorf("TestDictionaryDelete(%s): returned %v, expected %v", tst.input, returnVal, tst.expectedReturn)
		}
	}
}

func TestPackageDeclaration(t *testing.T) {
	tests := []struct{ input, expected string }{
		{``, `package(attr = "val")`},
		{`"""Docstring."""

load(":path.bzl", "x")

# package() comes here

x = 2`,
			`"""Docstring."""

load(":path.bzl", "x")

# package() comes here

package(attr = "val")

x = 2`,
		},
	}

	for _, tst := range tests {
		bld, err := build.Parse("BUILD", []byte(tst.input))
		if err != nil {
			t.Error(err)
			continue
		}
		pkg := PackageDeclaration(bld)
		pkg.SetAttr("attr", &build.StringExpr{Value: "val"})
		got := strings.TrimSpace(string(build.Format(bld)))
		want := strings.TrimSpace(tst.expected)

		if got != want {
			t.Errorf("TestPackageDeclaration: got:\n%s\nexpected:\n%s", got, want)
		}
	}
}
