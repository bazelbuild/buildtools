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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/tables"
	"github.com/bazelbuild/buildtools/testutils"
)

// exists reports whether the named file exists.
func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// setFlags sets flags based on the test file name (except the --type flag)
func setFlags(file string) func() {
	if strings.Contains(file, ".stripslashes.") {
		tables.StripLabelLeadingSlashes = true
	}
	// Test file 050 tests the ShortenAbsoluteLabelsToRelative behavior, all other tests assume that ShortenAbsoluteLabelsToRelative is false.
	if strings.Contains(file, string(os.PathSeparator)+"050.") {
		tables.ShortenAbsoluteLabelsToRelative = true
	}
	return func() {
		tables.StripLabelLeadingSlashes = false
		tables.ShortenAbsoluteLabelsToRelative = false
	}
}

func testIdempotence(t *testing.T, file string, isBuild bool) {
	defer setFlags(file)()
	testPrint(t, file, file, isBuild)
}

func testFormat(t *testing.T, input, output string, isBuild bool) {
	defer setFlags(output)()
	testPrint(t, input, output, isBuild)
}

// Test that reading and then writing the golden files
// does not change their output.
func TestPrintGolden(t *testing.T) {
	outs, chdir := findTests(t, ".golden")
	defer chdir()

	// Run the tests with --type=build
	for _, out := range outs {
		if strings.HasSuffix(out, ".bzl.golden") {
			continue
		}
		testIdempotence(t, out, true)
	}

	// Run the same tests with --type=bzl
	for _, out := range outs {
		if strings.HasSuffix(out, ".build.golden") {
			continue
		}
		testIdempotence(t, out, false)
	}
}

// Test that formatting the input files produces the golden files.
func TestPrintRewrite(t *testing.T) {
	ins, chdir := findTests(t, ".in")
	defer chdir()
	for _, in := range ins {
		prefix := in[:len(in)-len(".in")]
		outBzl := prefix + ".golden"
		outBuild := prefix + ".golden"

		if !exists(outBzl) {
			outBzl = prefix + ".bzl.golden"
			outBuild = prefix + ".build.golden"
		}

		testFormat(t, in, outBzl, false)
		testFormat(t, in, outBuild, true)

		stripslashesBuild := prefix + ".stripslashes.golden"
		if exists(stripslashesBuild) {
			// Test this file in BUILD mode only
			testFormat(t, in, stripslashesBuild, true)
		}
	}
}

// Test that golden files for the build mode aren't modified if reformatted in bzl mode
func TestPrintBuildAsBzl(t *testing.T) {
	ins, chdir := findTests(t, ".in")
	defer chdir()
	for _, in := range ins {
		prefix := in[:len(in)-len(".in")]
		outBuild := prefix + ".build.golden"
		if !exists(outBuild) {
			continue
		}
		testIdempotence(t, outBuild, false)
	}
}

// Test that golden files for the bzl become golden files for the build mode after reformatting
// in the build mode
func TestPrintBzlAsBuild(t *testing.T) {
	ins, chdir := findTests(t, ".in")
	defer chdir()
	for _, in := range ins {
		prefix := in[:len(in)-len(".in")]
		outBuild := prefix + ".build.golden"
		outBzl := prefix + ".bzl.golden"
		if !exists(outBuild) {
			continue
		}
		testFormat(t, outBzl, outBuild, true)
	}
}

// findTests finds all files of the passed suffix in the build/testdata directory.
// It changes the working directory to be the directory containing the `testdata` directory,
// and returns a function to call to change back to the current directory.
// This allows tests to assert on alias finding between absolute and relative labels.
func findTests(t *testing.T, suffix string) ([]string, func()) {
	return testutils.FindTests(t, "build", "testdata/*"+suffix)
}

// testPrint is a helper for testing the printer.
// It reads the file named in, reformats it, and compares
// the result to the file named out. If rewrite is true, the
// reformatting includes buildifier's higher-level rewrites.
func testPrint(t *testing.T, in, out string, isBuild bool) {
	data, err := ioutil.ReadFile(in)
	if err != nil {
		t.Error(err)
		return
	}

	golden, err := ioutil.ReadFile(out)
	if err != nil {
		t.Error(err)
		return
	}

	base := "testdata/" + filepath.Base(in)
	parser := ParseDefault
	if isBuild {
		parser = ParseBuild
	}

	file, err := parser(base, data)
	if err != nil {
		t.Error(err)
		return
	}

	Rewrite(file, nil)

	ndata := Format(file)

	if !bytes.Equal(ndata, golden) {
		t.Errorf("formatted %s incorrectly: diff shows -%s, +ours", base, filepath.Base(out))
		testutils.Tdiff(t, golden, ndata)
		return
	}
}

// Test that when files in the testdata directory are parsed
// and printed and parsed again, we get the same parse tree
// both times.
func TestPrintParse(t *testing.T) {
	outs, chdir := findTests(t, "")
	defer chdir()
	for _, out := range outs {
		data, err := ioutil.ReadFile(out)
		if err != nil {
			t.Error(err)
			continue
		}

		base := "testdata/" + filepath.Base(out)
		f, err := Parse(base, data)
		if err != nil {
			t.Errorf("parsing original: %v", err)
		}

		ndata := Format(f)

		f2, err := Parse(base, ndata)
		if err != nil {
			t.Errorf("parsing reformatted: %v", err)
		}

		eq := eqchecker{file: base}
		if err := eq.check(f, f2); err != nil {
			t.Errorf("not equal: %v", err)
		}
	}
}

// An eqchecker holds state for checking the equality of two parse trees.
type eqchecker struct {
	file string
	pos  Position
}

// errorf returns an error described by the printf-style format and arguments,
// inserting the current file position before the error text.
func (eq *eqchecker) errorf(format string, args ...interface{}) error {
	return fmt.Errorf("%s:%d: %s", eq.file, eq.pos.Line,
		fmt.Sprintf(format, args...))
}

// check checks that v and w represent the same parse tree.
// If not, it returns an error describing the first difference.
func (eq *eqchecker) check(v, w interface{}) error {
	return eq.checkValue(reflect.ValueOf(v), reflect.ValueOf(w))
}

var (
	posType        = reflect.TypeOf(Position{})
	commentsType   = reflect.TypeOf(Comments{})
	parenType      = reflect.TypeOf((*ParenExpr)(nil))
	stringExprType = reflect.TypeOf(StringExpr{})
)

// checkValue checks that v and w represent the same parse tree.
// If not, it returns an error describing the first difference.
func (eq *eqchecker) checkValue(v, w reflect.Value) error {
	// inner returns the innermost expression for v.
	// If v is a parenthesized expression (X) it returns x.
	// if v is a non-nil interface value, it returns the concrete
	// value in the interface.
	inner := func(v reflect.Value) reflect.Value {
		for v.IsValid() {
			if v.Type() == parenType {
				v = v.Elem().FieldByName("X")
				continue
			}
			if v.Kind() == reflect.Interface && !v.IsNil() {
				v = v.Elem()
				continue
			}
			break
		}
		return v
	}

	v = inner(v)
	w = inner(w)

	if v.Kind() != w.Kind() {
		return eq.errorf("%s became %s", v.Kind(), w.Kind())
	}

	// There is nothing to compare for zero values, so exit early.
	if !v.IsValid() {
		return nil
	}

	if v.Type() != w.Type() {
		return eq.errorf("%s became %s", v.Type(), w.Type())
	}

	if p, ok := v.Interface().(Expr); ok {
		eq.pos, _ = p.Span()
	}

	switch v.Kind() {
	default:
		return eq.errorf("unexpected type %s", v.Type())

	case reflect.Bool, reflect.Int, reflect.String:
		vi := v.Interface()
		wi := w.Interface()
		if vi != wi {
			return eq.errorf("%v became %v", vi, wi)
		}

	case reflect.Slice:
		vl := v.Len()
		wl := w.Len()
		for i := 0; i < vl || i < wl; i++ {
			if i >= vl {
				return eq.errorf("unexpected %s", w.Index(i).Type())
			}
			if i >= wl {
				return eq.errorf("missing %s", v.Index(i).Type())
			}
			if err := eq.checkValue(v.Index(i), w.Index(i)); err != nil {
				return err
			}
		}

	case reflect.Struct:
		// Fields in struct must match.
		t := v.Type()
		n := t.NumField()
		for i := 0; i < n; i++ {
			tf := t.Field(i)
			switch {
			default:
				if err := eq.checkValue(v.Field(i), w.Field(i)); err != nil {
					return err
				}

			case tf.Type == posType: // ignore positions
			case tf.Type == commentsType: // ignore comment assignment
			case tf.Name == "MultiLine": // ignore multiline setting
			case tf.Name == "LineBreak": // ignore line break setting
			case t == stringExprType && tf.Name == "Token": // ignore raw string token
			}
		}

	case reflect.Ptr, reflect.Interface:
		if v.IsNil() != w.IsNil() {
			if v.IsNil() {
				return eq.errorf("unexpected %s", w.Elem().Type())
			}
			return eq.errorf("missing %s", v.Elem().Type())
		}
		if err := eq.checkValue(v.Elem(), w.Elem()); err != nil {
			return err
		}
	}
	return nil
}
