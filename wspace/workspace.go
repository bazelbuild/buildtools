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

// Package wspace provides a method to find the root of the bazel tree.
package wspace

import (
	"os"
	"path/filepath"
)

const workspaceFile = "WORKSPACE"

// findContextPath finds the context path inside of a WORKSPACE-rooted source tree.
func findContextPath(rootDir string) (string, error) {
	if rootDir == "" {
		return os.Getwd()
	}
	return rootDir, nil
}

// FindWorkspaceRoot splits the current code context (the rootDir if present,
// the working directory if not.) It returns the path of the directory
// containing the WORKSPACE file, and the rest.
func FindWorkspaceRoot(rootDir string) (root string, rest string) {
	wd, err := findContextPath(rootDir)
	if err != nil {
		return "", ""
	}
	if root, err = Find(wd); err != nil {
		return "", ""
	}
	if len(wd) == len(root) {
		return root, ""
	}
	return root, wd[len(root)+1:]
}

// Find searches from the given dir and up for the WORKSPACE file
// returning the directory containing it, or an error if none found in the tree.
func Find(dir string) (string, error) {
	if dir == "" || dir == "/" || dir == "." {
		return "", os.ErrNotExist
	}
	if _, err := os.Stat(filepath.Join(dir, workspaceFile)); err == nil {
		return dir, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return Find(filepath.Dir(dir))
}
