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
	"os"
	"path"
	"path/filepath"
	"strings"
)

// containsFile checks whether a given directory contains a file called
// <file> or <file>.bazel.
func containsFile(dir, file string) bool {
	for _, filename := range []string{file, file + ".bazel"} {
		path := filepath.Join(dir, filename)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

// SplitRelativePath splits a path relative to the workspace root into package name and label.
// It takes the workspace root and chunks of the relative path (already split) as input arguments.
// Both output variables always have forward slashes as separators.
func SplitRelativePath(workspaceRoot string, chunks []string) (pkg, label string) {
	switch chunks[len(chunks)-1] {
	case "BUILD", "BUILD.bazel":
		return path.Join(chunks[:len(chunks)-1]...), chunks[len(chunks)-1]
	}

	pkg = ""
	label = path.Join(chunks...)
	parent := workspaceRoot
	for i, chunk := range chunks {
		if i == len(chunks)-1 {
			// The last chunk is a filename, not a directory
			break
		}
		parent = filepath.Join(parent, chunk)
		if containsFile(parent, "BUILD") {
			pkg = path.Join(chunks[:i+1]...)
			label = path.Join(chunks[i+1:]...)
		}
	}
	return pkg, label
}

// SplitFilePath splits a file path into the workspace root, package name and label.
// Workspace root is determined as the last directory in the file path that
// contains a WORKSPACE (or WORKSPACE.bazel) file.
// Package and label are always separated with forward slashes.
// Returns empty strings if no WORKSPACE file is found.
func SplitFilePath(filename string) (workspaceRoot, pkg, label string) {
	root := "/"
	if volume := filepath.VolumeName(filename); volume != "" {
		// Windows
		root = volume + "\\"
	}
	// filename relative to the file system root
	relPath := filename[len(root):]

	chunks := append([]string{""}, strings.Split(relPath, string(os.PathSeparator))...)
	parent := root
	workspaceIndex := -1
	for i, chunk := range chunks {
		if i == len(chunks)-1 {
			// The last chunk is a filename, not a directory
			break
		}
		parent = filepath.Join(parent, chunk)
		if containsFile(parent, "WORKSPACE") {
			workspaceRoot = parent
			workspaceIndex = i
		}
	}
	if workspaceIndex != -1 {
		pkg, label = SplitRelativePath(workspaceRoot, chunks[workspaceIndex+1:])
	}
	return workspaceRoot, pkg, label
}
