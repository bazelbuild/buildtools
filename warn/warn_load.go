package warn

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/tables"
)

func ruleLoadLocationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	for stmtIndex := 0; stmtIndex < len(f.Stmt); stmtIndex++ {
		load, ok := f.Stmt[stmtIndex].(*build.LoadStmt)
		if !ok {
			continue
		}

		for i := 0; i < len(load.To); i++ {
			from := load.From[i]

			expectedLocation, ok := tables.RuleLoadLocation[from.Name]
			if !ok || expectedLocation == load.Module.Value {
				continue
			}

			f := makeLinterFinding(load.From[i], fmt.Sprintf("Rule %q must be loaded from %v.", from.Name, expectedLocation))
			findings = append(findings, f)
		}

	}
	return findings
}
