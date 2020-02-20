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

		pkg := "//" + f.Pkg // Canonical name of the package
		chunks := internalDirectory.Split(load.Module.Value, 2)
		if len(chunks) < 2 {
			continue
		}

		if strings.HasPrefix(pkg, chunks[0]) {
			continue
		}

		findings = append(findings, makeLinterFinding(
			load.Module,
			fmt.Sprintf("Module %q can only be loaded from files located inside %q, not from %q.", load.Module.Value, chunks[0], pkg)))
	}

	return findings
}
