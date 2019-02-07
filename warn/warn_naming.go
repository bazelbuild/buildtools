package warn

import (
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
)

var ambiguousNames = map[string]bool{
	"I": true,
	"l": true,
	"O": true,
}

// ambiguousNameCheck checks for the names of idents and functions
func ambiguousNameCheck(f *build.File, expr build.Expr, findings []*Finding) []*Finding {
	var name string
	switch expr := expr.(type) {
	case *build.Ident:
		name = expr.Name
	case *build.DefStmt:
		name = expr.Name
	default:
		return findings
	}

	if ambiguousNames[name] {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "confusing-name",
				`Never use 'l', 'I', or 'O' as names (they're too easily confused with 'I', 'l', or '0').`, true, nil))
	}
	return findings
}

func confusingNameWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	// check for global variable names
	for _, ident := range collectLocalVariables(f.Stmt) {
		findings = ambiguousNameCheck(f, ident, findings)
	}

	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.DefStmt:
			findings = ambiguousNameCheck(f, expr, findings)
			for _, param := range getFunctionParams(expr) {
				findings = ambiguousNameCheck(f, param, findings)
			}
			for _, ident := range collectLocalVariables(expr.Body) {
				findings = ambiguousNameCheck(f, ident, findings)
			}
		case *build.Comprehension:
			for _, clause := range expr.Clauses {
				forClause, ok := clause.(*build.ForClause)
				if !ok {
					continue
				}
				for _, ident := range bzlenv.CollectLValues(forClause.Vars) {
					findings = ambiguousNameCheck(f, ident, findings)
				}
			}
		}
	})

	return findings
}
