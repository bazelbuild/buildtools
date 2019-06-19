// Warnings about deprecated Bazel-related operations

package warn

import "github.com/bazelbuild/buildtools/build"

func depsetUnionWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding
	addWarning := func(expr build.Expr) {
		findings = append(findings,
			makeLinterFinding(expr, `Depsets should be joined using the "depset()" constructor.`))
	}

	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.BinaryExpr:
			// `depset1 + depset2` or `depset1 | depset2`
			if types[expr.X] != Depset && types[expr.Y] != Depset {
				return
			}
			switch expr.Op {
			case "+", "|":
				addWarning(expr)
			}
		case *build.AssignExpr:
			// `depset1 += depset2` or `depset1 |= depset2`
			if types[expr.LHS] != Depset && types[expr.RHS] != Depset {
				return
			}
			switch expr.Op {
			case "+=", "|=":
				addWarning(expr)
			}
		case *build.CallExpr:
			// `depset1.union(depset2)`
			if len(expr.List) == 0 {
				return
			}
			dot, ok := expr.X.(*build.DotExpr)
			if !ok {
				return
			}
			if dot.Name != "union" {
				return
			}
			if types[dot.X] != Depset && types[expr.List[0]] != Depset {
				return
			}
			addWarning(expr)
		}
	})
	return findings
}

func depsetIterationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	addFinding := func(expr *build.Expr) {
		_, end := (*expr).Span()
		newNode := &build.CallExpr{
			X: &build.DotExpr{
				X:    *expr,
				Name: "to_list",
			},
			End: build.End{Pos: end},
		}
		findings = append(findings,
			makeLinterFinding(*expr, `Depset iteration is deprecated, use the "to_list()" method instead.`, LinterReplacement{expr, newNode}))
	}

	types := detectTypes(f)
	build.WalkPointers(f, func(e *build.Expr, stack []build.Expr) {
		switch expr := (*e).(type) {
		case *build.ForStmt:
			if types[expr.X] != Depset {
				return
			}
			addFinding(&expr.X)
		case *build.ForClause:
			if types[expr.X] != Depset {
				return
			}
			addFinding(&expr.X)
		case *build.BinaryExpr:
			if expr.Op != "in" && expr.Op != "not in" {
				return
			}
			if types[expr.Y] != Depset {
				return
			}
			addFinding(&expr.Y)
		case *build.CallExpr:
			ident, ok := expr.X.(*build.Ident)
			if !ok {
				return
			}
			switch ident.Name {
			case "all", "any", "depset", "len", "sorted", "max", "min", "list", "tuple":
				if len(expr.List) != 1 {
					return
				}
				if types[expr.List[0]] != Depset {
					return
				}
				addFinding(&expr.List[0])
				if ident.Name == "list" {
					// `list(d.to_list())` can be simplified to just `d.to_list()`
					findings[len(findings)-1].Replacement[0].Old = e
				}
			case "zip":
				for i, arg := range expr.List {
					if types[arg] != Depset {
						continue
					}
					addFinding(&expr.List[i])
				}
			}
		}
		return
	})
	return findings
}
