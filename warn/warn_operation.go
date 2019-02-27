// Warnings about deprecated operations in Starlark

package warn

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
)

func dictionaryConcatenationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	var addWarning = func(expr build.Expr) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "dict-concatenation",
				"Dictionary concatenation is deprecated.", true, nil))
	}

	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.BinaryExpr:
			if expr.Op != "+" {
				return
			}
			if types[expr.X] == Dict || types[expr.Y] == Dict {
				addWarning(expr)
			}
		case *build.AssignExpr:
			if expr.Op != "+=" {
				return
			}
			if types[expr.LHS] == Dict || types[expr.RHS] == Dict {
				addWarning(expr)
			}
		}
	})
	return findings
}

func stringIterationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	addWarning := func(expr build.Expr) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "string-iteration",
				"String iteration is deprecated.", true, nil))
	}

	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.ForStmt:
			if types[expr.X] == String {
				addWarning(expr.X)
			}
		case *build.ForClause:
			if types[expr.X] == String {
				addWarning(expr.X)
			}
		case *build.CallExpr:
			ident, ok := expr.X.(*build.Ident)
			if !ok {
				return
			}
			switch ident.Name {
			case "all", "any", "reversed", "max", "min":
				if len(expr.List) != 1 {
					return
				}
				if types[expr.List[0]] == String {
					addWarning(expr.List[0])
				}
			case "zip":
				for _, arg := range expr.List {
					if types[arg] == String {
						addWarning(arg)
					}
				}
			}
		}
	})
	return findings
}

func integerDivisionWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	var addWarning = func(expr build.Expr, op string) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "integer-division",
				fmt.Sprintf(`The "%s" operator for integer division is deprecated in favor of "/%s".`, op, op), true, nil))
	}

	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.BinaryExpr:
			if expr.Op != "/" {
				return
			}
			if fix {
				expr.Op = "//"
				return
			}
			addWarning(expr, expr.Op)

		case *build.AssignExpr:
			if expr.Op != "/=" {
				return
			}
			if fix {
				expr.Op = "//="
				return
			}
			addWarning(expr, expr.Op)
		}
	})
	return findings
}
