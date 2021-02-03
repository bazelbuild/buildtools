/*
Copyright 2021 Google LLC

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

package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bazelbuild/buildtools/testutils"
	"github.com/bazelbuild/buildtools/warn"
)

func TestAllWarningsAreDocumented(t *testing.T) {
	testdata := path.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))

	textprotoPath := path.Join(testdata, "warn", "docs", "warnings.textproto")
	warnings, err := readWarningsFromFile(textprotoPath)
	if err != nil {
		t.Fatalf("getWarnings(%q) = %v", textprotoPath, err)
	}

	names := make(map[string]bool)
	for _, w := range warnings.Warnings {
		for _, n := range w.Name {
			names[n] = true
		}
	}

	for _, n := range warn.AllWarnings {
		if !names[n] {
			t.Errorf("warning %q is not documented in warn/docs/warnings.textproto", n)
		}
	}
}

func TestFilesMatch(t *testing.T) {
	testdata := path.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))

	generatedPath := path.Join(testdata, "warn", "docs", "WARNINGS.md") 
	generated, err := ioutil.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v", generatedPath, err)
	}

	checkedInPath := path.Join(testdata, "WARNINGS.md")
	checkedIn, err := ioutil.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v", checkedInPath, err)
	}

	d, err := testutils.Diff(generated, checkedIn)
	if err != nil {
		t.Fatalf("diff(generated, checkedIn) returns error:", err)
	}
	if len(d) != 0 {
		t.Errorf("diff(generated, checkedIn) = %v\n", string(d))
		t.Errorf("To update the documentation, run `bazel build //warn/docs:warnings_docs && cp bazel-bin/warn/docs/WARNINGS.md .`")
	}
}
