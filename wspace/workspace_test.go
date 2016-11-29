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
package wspace

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type testCase struct {
	input                      string
	expectedRoot, expectedRest string
}

func TestBasic(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	if err := os.MkdirAll(filepath.Join(tmp, "a", "b", "c"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmp, workspaceFile), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmp, "a", "b", workspaceFile), nil, 0755); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []testCase{
		{tmp, tmp, ""},
		{filepath.Join(tmp, "a"), tmp, "a"},
		{filepath.Join(tmp, "a", "b"), filepath.Join(tmp, "a", "b"), ""},
		{filepath.Join(tmp, "a", "b", "c"), filepath.Join(tmp, "a", "b"), "c"},
		{"a", "", ""}, // error case
	} {
		root, rest := FindWorkspaceRoot(tc.input)
		if root != tc.expectedRoot || rest != tc.expectedRest {
			t.Errorf("FindWorkspaceRoot(%q) = %q, %q; want %q, %q", tc.input, root, rest, tc.expectedRoot, tc.expectedRest)
		}
	}
}
