package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIsStarlarkFile(t *testing.T) {
	tests := []struct {
		filename string
		ok       bool
	}{
		{
			filename: "BUILD",
			ok:       true,
		},
		{
			filename: "BUILD.bazel",
			ok:       true,
		},
		{
			filename: "BUILD.oss",
			ok:       true,
		},
		{
			filename: "BUILD.foo.bazel",
			ok:       true,
		},
		{
			filename: "BUILD.foo.oss",
			ok:       true,
		},
		{
			filename: "build.foo.bazel",
			ok:       false,
		},
		{
			filename: "build.foo.oss",
			ok:       false,
		},
		{
			filename: "build.oss",
			ok:       false,
		},
		{
			filename: "WORKSPACE",
			ok:       true,
		},
		{
			filename: "WORKSPACE.bazel",
			ok:       true,
		},
		{
			filename: "WORKSPACE.oss",
			ok:       true,
		},
		{
			filename: "WORKSPACE.foo.bazel",
			ok:       true,
		},
		{
			filename: "WORKSPACE.foo.oss",
			ok:       true,
		},
		{
			filename: "workspace.foo.bazel",
			ok:       false,
		},
		{
			filename: "workspace.foo.oss",
			ok:       false,
		},
		{
			filename: "workspace.oss",
			ok:       false,
		},
		{
			filename: "build.gradle",
			ok:       false,
		},
		{
			filename: "workspace.xml",
			ok:       false,
		},
		{
			filename: "foo.bzl",
			ok:       true,
		},
		{
			filename: "foo.BZL",
			ok:       false,
		},
		{
			filename: "build.bzl",
			ok:       true,
		},
		{
			filename: "workspace.sky",
			ok:       true,
		},
		{
			filename: "foo.bar",
			ok:       false,
		},
		{
			filename: "foo.build",
			ok:       false,
		},
		{
			filename: "foo.workspace",
			ok:       false,
		},
	}

	for _, tc := range tests {
		if isStarlarkFile(tc.filename) != tc.ok {
			t.Errorf("Wrong result for %q, want %t", tc.filename, tc.ok)
		}
	}
}

func checkSplitFilePathOutput(t *testing.T, name, filename, expectedWorkspaceRoot, expectedPkg string) {
	workspaceRoot, pkg := SplitFilePath(filename)
	if workspaceRoot != expectedWorkspaceRoot {
		t.Errorf("%s: expected the workspace root to be %q, was %q instead", name, expectedWorkspaceRoot, workspaceRoot)
	}
	if pkg != expectedPkg {
		t.Errorf("%s: expected the package name to be %q, was %q instead", name, expectedPkg, pkg)
	}
}

func TestSplitFilePath(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "path", "to", "package", "file.bzl")
	checkSplitFilePathOutput(t, "No WORKSPACE file", filename, "", "")

	// Create a WORKSPACE file and try again (dir/WORKSPACE)
	if err := ioutil.WriteFile(filepath.Join(dir, "WORKSPACE"), []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE file exists", filename, dir, "path/to/package")
	checkSplitFilePathOutput(t, "WORKSPACE file exists, empty package", filepath.Join(dir, "file.bzl"), dir, "")

	// Rename WORKSPACE to WORKSPACE.bazel and try again (dir/WORKSPACE.bazel)
	if err := os.Rename(filepath.Join(dir, "WORKSPACE"), filepath.Join(dir, "WORKSPACE.bazel")); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "WORKSPACE file exists", filename, dir, "path/to/package")
	checkSplitFilePathOutput(t, "WORKSPACE file exists, empty package", filepath.Join(dir, "file.bzl"), dir, "")

	// Create another WORKSPACE file and try again (dir/path/WORKSPACE)
	newRoot := filepath.Join(dir, "path")
	if err := os.MkdirAll(newRoot, os.ModePerm); err != nil {
		t.Error(err)
	}
	if err := ioutil.WriteFile(filepath.Join(newRoot, "WORKSPACE"), []byte{}, os.ModePerm); err != nil {
		t.Error(err)
	}
	checkSplitFilePathOutput(t, "Two WORKSPACE files exist", filename, newRoot, "to/package")
}
