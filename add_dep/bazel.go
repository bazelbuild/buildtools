/*
Copyright 2024 Google LLC

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

// Package bazel contains types representing Blaze concepts, such as packages and rules.
package bazel

import (
	"fmt"
	"path"
	"strings"
)

// Label is a rule's label, e.g. //foo:bar.
type Label string

// Split splits a label to its package name and rule name parts.
// Example: //foo:bar --> "foo", "bar"
func (l Label) Split() (pkgName, ruleName string) {
	s := strings.TrimPrefix(string(l), "//")
	i := strings.IndexByte(s, ':')
	if i == -1 {
		return s, s[strings.LastIndexByte(s, '/')+1:]
	}
	return s[:i], s[i+1:]
}

// ParseRelativeLabel parses a label, not necessarily absolute,
// relative to some package.
//
// If the label is absolute, for instance "//a/b" or "//a/b:c", the
// result is the same as ParseAbsoluteLabel.  If the label is relative,
// for instance ":foo" or "bar.go", the result is the same as
// ParseAbsoluteLabel on "//pkg:foo" or "//pkg:bar.go", respectively.
func ParseRelativeLabel(pkg, s string) (Label, error) {
	if strings.HasPrefix(s, "//") || strings.HasPrefix(s, "@") {
		return ParseAbsoluteLabel(s)
	}
	if strings.Count(pkg, "//") > 1 {
		return "", fmt.Errorf("package name %q contains '//' more than once", pkg)
	}
	if s == "" {
		return "", fmt.Errorf("empty label")
	}
	colonIdx := strings.Index(s, ":")
	if colonIdx > 0 {
		return "", fmt.Errorf("label %q doesn't start with // or @, but also contains a colon", s)
	}
	if s[0] == ':' {
		s = s[1:]
	}
	if strings.HasPrefix(pkg, "@") {
		return Label(pkg + ":" + s), nil
	}
	return Label("//" + pkg + ":" + s), nil
}

// ParseAbsoluteLabel parses a label string in absolute form, such as
// "//aaa/bbb:ccc/ddd" or "//aaa/bbb".
//
// See https://bazel.build/versions/master/docs/build-ref.html#labels,
func ParseAbsoluteLabel(s string) (Label, error) {
	if !strings.HasPrefix(s, "//") && !strings.HasPrefix(s, "@") {
		return "", fmt.Errorf("absolute label must start with // or @, %q is neither", s)
	}
	i := strings.Index(s, "//")
	if i < 0 {
		return "", fmt.Errorf("invalid label %q", s)
	}
	// TODO(b/36533053): Bazel accepts invalid labels starting with more than two slashes,
	// thus so must we for now.
	s = strings.TrimLeft(s, "/")

	var pkg, name string
	if i = strings.IndexByte(s, ':'); i < 0 {
		// "//foo/bar"
		pkg = s
		name = path.Base(s)
	} else {
		// "//foo/bar:wiz"
		pkg = s[:i]
		name = s[i+1:]
	}
	if strings.Count(pkg, "//") > 1 {
		return "", fmt.Errorf("package name %q contains '//' more than once", pkg)
	}
	if strings.Index(name, "//") != -1 {
		return "", fmt.Errorf("target name %q contains '//'", pkg)
	}
	if strings.Index(name, ":") != -1 {
		return "", fmt.Errorf("target name %q contains ':'", pkg)
	}

	if strings.HasPrefix(pkg, "@") {
		return Label(pkg + ":" + name), nil
	}
	return Label("//" + pkg + ":" + name), nil
}
