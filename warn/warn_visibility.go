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

// Warnings about visibility of .bzl files

package warn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

var internalDirectory = regexp.MustCompile("/(internal|private)[/:]")

func bzlVisibilityWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	if f.WorkspaceRoot == "" {
		// Empty workspace root means buildifier doesn't know the location of
		// the file relative to the workspace directory and can't warn about .bzl
		// file visibility correctly.
		return findings
	}

	for _, stmt := range f.Stmt {
		load, ok := stmt.(*build.LoadStmt)
		if !ok || load.Module == nil {
			continue
		}

		// A load statement may use a fully qualified module name (including a
		// repository name). Buildifier should check if the repository name refers
		// to the current repository, but because it doesn't know the current
		// repository name it's better to assume that it matches the repository name
		// in the load statement: this way it may miss some usages of private .bzl
		// files that aren't supposed to be visible, but won't show false-positive
		// warnings in case the private file is actually allowed to be used.
		module := load.Module.Value
		if strings.HasPrefix(module, "@") {
			if chunks := strings.SplitN(module, "//", 2); len(chunks) == 2 {
				module = "//" + chunks[1]
			}
		}

		path := f.CanonicalPath() // Canonical name of the file
		chunks := internalDirectory.Split(module, 2)
		if len(chunks) < 2 {
			continue
		}

		if strings.HasPrefix(path, chunks[0]) ||
			strings.HasPrefix(strings.Replace(path, "/javatest/", "/java/", 1), chunks[0]) {
			continue
		}

		findings = append(findings, makeLinterFinding(
			load.Module,
			fmt.Sprintf("Module %q can only be loaded from files located inside %q, not from %q.", load.Module.Value, chunks[0], path)))
	}

	return findings
}
