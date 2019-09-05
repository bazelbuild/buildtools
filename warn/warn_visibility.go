// Warnings about visibility of .bzl files

package warn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

var internalDirectory = regexp.MustCompile("/(internal|private)[/:]")

func deprecatedBzlLoadWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

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
