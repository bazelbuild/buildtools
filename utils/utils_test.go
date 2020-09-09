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

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func checkSplitFilePathOutput(t *testing.T, name, filename, expectedWorkspaceRoot, expectedPkg, expectedLabel string) {
	workspaceRoot, pkg, label := SplitFilePath(filename)
	if workspaceRoot != expectedWorkspaceRoot {
		t.Errorf("%s: expected the workspace root to be %q, was %q instead", name, expectedWorkspaceRoot, workspaceRoot)
	}
	if pkg != expectedPkg {
		t.Errorf("%s: expected the package name to be %q, was %q instead", name, expectedPkg, pkg)
	}
	if label != expectedLabel {
		t.Errorf("%s: expected the label to be %q, was %q instead", name, expectedLabel, label)
	}
}

func TestSplitFilePath(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(filepath.Join(dir, "path", "to", "package"), os.ModePerm); err != nil {
		t.Error(err)
	}

	filename := filepath.Join(dir, "path", "to", "package", "file.bzl")
	checkSplitFilePathOutput(t, "No WORKSPACE file", filename, "", "", "")

	// Create a WORKSPACE file and try again (dir/WORKSPACE)
	if err := ioutil.WriteFile(filepath.Join(dir, "WORKSPACE"), []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE file exists", filename, dir, "", "path/to/package/file.bzl")
	checkSplitFilePathOutput(t, "WORKSPACE file exists, empty package", filepath.Join(dir, "file.bzl"), dir, "", "file.bzl")

	// Add a BUILD file
	buildPath := filepath.Join(dir, "path", "to", "BUILD")
	if err := ioutil.WriteFile(buildPath, []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE and BUILD files exists 1", buildPath, dir, "path/to", "BUILD")
	checkSplitFilePathOutput(t, "WORKSPACE and BUILD files exists 2", filename, dir, "path/to", "package/file.bzl")

	// Add a subpackage BUILD file
	subBuildPath := filepath.Join(dir, "path", "to", "package", "BUILD.bazel")
	if err := ioutil.WriteFile(subBuildPath, []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE and two BUILD files exists 1", subBuildPath, dir, "path/to/package", "BUILD.bazel")
	checkSplitFilePathOutput(t, "WORKSPACE and two BUILD files exists 2", filename, dir, "path/to/package", "file.bzl")

	// Rename WORKSPACE to WORKSPACE.bazel and try again (dir/WORKSPACE.bazel)
	if err := os.Rename(filepath.Join(dir, "WORKSPACE"), filepath.Join(dir, "WORKSPACE.bazel")); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE file exists", filename, dir, "path/to/package", "file.bzl")
	checkSplitFilePathOutput(t, "WORKSPACE file exists, empty package", filepath.Join(dir, "file.bzl"), dir, "", "file.bzl")

	// Create another WORKSPACE file and try again (dir/path/WORKSPACE)
	newRoot := filepath.Join(dir, "path")
	if err := os.MkdirAll(newRoot, os.ModePerm); err != nil {
		t.Error(err)
	}
	if err := ioutil.WriteFile(filepath.Join(newRoot, "WORKSPACE"), []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "Two WORKSPACE files exist", filename, newRoot, "to/package", "file.bzl")
}
