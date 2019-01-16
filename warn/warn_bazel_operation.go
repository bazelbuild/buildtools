// Warnings about deprecated Bazel-related operations

package warn

import "github.com/bazelbuild/buildtools/build"

func depsetUnionWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	addWarning := func(expr build.Expr) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "depset-union",
				"Depsets should be joined using the depset constructor.", true, nil))
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
			case "+", "|", "+=", "|=":
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

func depsetIterationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	addWarning := func(expr build.Expr) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "depset-iteration",
				"Depset iteration is deprecated.", true, nil))
	}

	// fixNode returns a call for .to_list() on the input node (assuming that it's a depset)
	fixNode := func(expr build.Expr) build.Expr {
		_, end := expr.Span()
		return &build.CallExpr{
			X: &build.DotExpr{
				X:    expr,
				Name: "to_list",
			},
			End: build.End{Pos: end},
		}
	}

	types := detectTypes(f)
	build.Edit(f, func(expr build.Expr, stack []build.Expr) build.Expr {
		switch expr := expr.(type) {
		case *build.ForStmt:
			if types[expr.X] != Depset {
				return nil
			}
			if !fix {
				addWarning(expr.X)
				return nil
			}
			expr.X = fixNode(expr.X)
		case *build.ForClause:
			if types[expr.X] != Depset {
				return nil
			}
			if !fix {
				addWarning(expr.X)
				return nil
			}
			expr.X = fixNode(expr.X)
		case *build.BinaryExpr:
			if expr.Op != "in" && expr.Op != "not in" {
				return nil
			}
			if types[expr.Y] != Depset {
				return nil
			}
			if !fix {
				addWarning(expr.Y)
				return nil
			}
			expr.Y = fixNode(expr.Y)
		case *build.CallExpr:
			ident, ok := expr.X.(*build.Ident)
			if !ok {
				return nil
			}
			switch ident.Name {
			case "all", "any", "depset", "len", "sorted", "max", "min", "list", "tuple":
				if len(expr.List) != 1 {
					return nil
				}
				if types[expr.List[0]] != Depset {
					return nil
				}
				if !fix {
					addWarning(expr.List[0])
					return nil
				}
				newNode := fixNode(expr.List[0])
				if ident.Name != "list" {
					expr.List[0] = newNode
					return nil
				}
				// `list(d.to_list())` can be simplified to just `d.to_list()`
				return newNode
			case "zip":
				for i, arg := range expr.List {
					if types[arg] != Depset {
						continue
					}
					if !fix {
						addWarning(arg)
						return nil
					}
					expr.List[i] = fixNode(arg)
				}
			}
		}
		return nil
	})
	return findings
}
