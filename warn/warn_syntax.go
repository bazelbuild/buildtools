// Warnings about potential syntax errors

package warn

import (
	"sort"

	"github.com/bazelbuild/buildtools/build"
)

func argumentsOrderWarning(f *build.File, fix bool) []*Finding {
	argumentType := func(expr build.Expr) int {
		switch expr := expr.(type) {
		case *build.UnaryExpr:
			switch expr.Op {
			case "**":
				return 4
			case "*":
				return 3
			}
		case *build.BinaryExpr:
			if expr.Op == "=" {
				return 2
			}
		}
		return 1
	}

	getComparator := func(args []build.Expr) func(i, j int) bool {
		return func(i, j int) bool {
			return argumentType(args[i]) < argumentType(args[j])
		}
	}

	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		comparator := getComparator(call.List)
		if fix {
			sort.SliceStable(call.List, comparator)
			return
		}
		if sort.SliceIsSorted(call.List, comparator) {
			return
		}
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "args-order",
				"Function call arguments should be in the following order: positional, keyword, *args, **kwargs.", true, nil))
	})
	return findings
}
