package warn

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
	"strings"
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

// isProviderNode returns whether the node is a call of `provider()`
func isProviderNode(expr build.Expr) bool {
	call, ok := expr.(*build.CallExpr)
	if !ok {
		return false
	}
	name, ok := call.X.(*build.Ident)
	return ok && name.Name == "provider"
}

func isUpperCamelCase(name string) bool {
	if strings.HasPrefix(name, "_") {
		// Private providers are allowed
		name = name[1:]
	}
	return !strings.ContainsRune(name, '_') && name == strings.Title(name)
}

func isLowerSnakeCase(name string) bool {
	return name == strings.ToLower(name)
}

func isUpperSnakeCase(name string) bool {
	return name == strings.ToUpper(name)
}

func nameConventionsWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	build.WalkStatements(f, func(stmt build.Expr, stack []build.Expr) {
		// looking for provider declaration statements: `xxx = provider()`
		// note that the code won't trigger on complex assignments, such as `x, y = foo, provider()`
		binary, ok := stmt.(*build.BinaryExpr)
		if !ok || binary.Op != "=" {
			return
		}
		if isProviderNode(binary.Y) {
			ident, ok := binary.X.(*build.Ident)
			if !ok {
				return
			}
			if !isUpperCamelCase(ident.Name) || !strings.HasSuffix(ident.Name, "Info") {
				start, end := ident.Span()
				findings = append(findings,
					makeFinding(f, start, end, "name-conventions",
						fmt.Sprintf(`Provider name "%s" should be UpperCamelCase and should end with 'Info'.`, ident.Name), true, nil))
			}
			return
		}
		for _, ident := range bzlenv.CollectLValues(binary.X) {
			if !isLowerSnakeCase(ident.Name) && !isUpperSnakeCase(ident.Name) {
				start, end := ident.Span()
				findings = append(findings,
					makeFinding(f, start, end, "name-conventions",
						fmt.Sprintf(`Variable name "%s" should be lower_snake_case or UPPER_SNAKE_CASE (for constants).`, ident.Name), true, nil))
			}
		}
	})

	return findings
}
