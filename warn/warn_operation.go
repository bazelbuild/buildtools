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

// Warnings about deprecated operations in Starlark

package warn

import (
	"github.com/bazelbuild/buildtools/build"
)

func dictionaryConcatenationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	var addWarning = func(expr build.Expr) {
		findings = append(findings,
			makeLinterFinding(expr, "Dictionary concatenation is deprecated."))
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

func stringIterationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	addWarning := func(expr build.Expr) {
		findings = append(findings,
			makeLinterFinding(expr, "String iteration is deprecated."))
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

func integerDivisionWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	build.WalkPointers(f, func(e *build.Expr, stack []build.Expr) {
		switch expr := (*e).(type) {
		case *build.BinaryExpr:
			if expr.Op != "/" {
				return
			}
			newBinary := *expr
			newBinary.Op = "//"
			findings = append(findings,
				makeLinterFinding(expr, `The "/" operator for integer division is deprecated in favor of "//".`,
					LinterReplacement{e, &newBinary}))

		case *build.AssignExpr:
			if expr.Op != "/=" {
				return
			}
			newAssign := *expr
			newAssign.Op = "//="
			findings = append(findings,
				makeLinterFinding(expr, `The "/=" operator for integer division is deprecated in favor of "//=".`,
					LinterReplacement{e, &newAssign}))
		}
	})
	return findings
}

func listAppendWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	build.WalkPointers(f, func(expr *build.Expr, stack []build.Expr) {
		as, ok := (*expr).(*build.AssignExpr)
		if !ok || as.Op != "+=" {
			return
		}

		list, ok := as.RHS.(*build.ListExpr)
		if !ok || len(list.List) != 1 {
			return
		}

		_, end := as.Span()
		findings = append(findings, makeLinterFinding(as, `Prefer using ".append()" to adding a single element list.`,
			LinterReplacement{expr, &build.CallExpr{
				Comments: as.Comments,
				X: &build.DotExpr{
					X:    as.LHS,
					Name: "append",
				},
				List: list.List,
				End:  build.End{Pos: end},
			}}))

	})
	return findings
}
