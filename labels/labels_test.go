/*
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package labels

import (
	"testing"
)

var parseLabelTests = []struct {
	in     string
	repo   string
	pkg    string
	target string
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
		l := Parse(tt.in)
		if l.Repository != tt.repo || l.Package != tt.pkg || l.Target != tt.target {
			t.Errorf("%d. Parse(%q) => (%q, %q, %q), want (%q, %q, %q)",
				i, tt.in, l.Repository, l.Package, l.Target, tt.repo, tt.pkg, tt.target)
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
	{"\"//baz\"", "", "\"//baz\""},
}

func TestShortenLabel(t *testing.T) {
	for i, tt := range shortenLabelTests {
		result := Shorten(tt.in, tt.pkg)
		if result != tt.result {
			t.Errorf("%d. Shorten(%q, %q) => %q, want %q",
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
		if got := Equal(tt.label1, tt.label2, tt.pkg); got != tt.expected {
			t.Errorf("%d. Equal(%q, %q, %q) => %v, want %v",
				i, tt.label1, tt.label2, tt.pkg, got, tt.expected)
		}
	}
}
