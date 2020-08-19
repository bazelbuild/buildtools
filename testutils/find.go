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

package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

// FindTests finds all files matching the given pattern.
// It changes the working directory to `directory`,  and returns a function
// to call to change back to the current directory.
// This allows tests to assert on alias finding between absolute and relative labels.
func FindTests(t *testing.T, directory, pattern string) ([]string, func()) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), directory)); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("Didn't find any test cases")
	}
	return files, func() { os.Chdir(wd) }
}
