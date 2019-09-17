// Warnings about deprecated operations in Starlark

package warn

import (
	"fmt"
	"strings"

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

func stringEscapeWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	build.WalkPointers(f, func(expr *build.Expr, stack []build.Expr) {
		str, ok := (*expr).(*build.StringExpr)
		if !ok || len(str.Token) == 0 || strings.HasPrefix(str.Token, "r") {
			// String literals with empty Token field may appear if they are manually created StringExpr nodes
			// (token is only used as a hint to the printer, if it doesn't exist, the Value field is used to
			// generate the string literal).
			// Raw strings are allowed to have backslashes anywhere.
			return
		}

		var quotes, value string
		if str.TripleQuote {
			quotes = str.Token[:3]
			value = str.Token[3 : len(str.Token)-3]
		} else {
			quotes = str.Token[:1]
			value = str.Token[1 : len(str.Token)-1]
		}

		var problems []int // positions of the problems (unidentified escape sequences)

		escaped := false
		// This for-loop doesn't correctly check for a backlash at the end of the string literal, but
		// such string can't be parsed anyway, neither by Bazel nor by Buildifier.
		for i, ch := range value {
			if !escaped {
				if ch == '\\' {
					escaped = true
				}
				continue
			}

			switch ch {
			case '\n', '\\', 'n', 'r', 't', 'x', '\'', '"', '0', '1', '2', '3', '4', '5', '6', '7':
				// According to https://github.com/Quarz0/bazel/blob/207a6103393908aba64ddb96239fbdd56cdfec05/src/main/java/com/google/devtools/build/lib/syntax/Lexer.java
				// \x is also included to the list, although it's not supported by Bazel, but it's supported
				// by Buildifier. This is safe for the migration because it's never been supported in Bazel,
				// even before --incompatible_restrict_string_escapes was flipped.
			default:
				problems = append(problems, i)
			}
			escaped = false
		}

		if len(problems) == 0 {
			return
		}

		var msg string
		if len(problems) == 1 {
			msg = fmt.Sprintf(
				"Invalid escape sequence \\%s at position %d.",
				string(value[problems[0]]),
				problems[0],
			)
		} else {
			var builder strings.Builder
			builder.WriteString("Invalid escape sequences:\n")
			for _, pos := range problems {
				builder.WriteString(fmt.Sprintf(
					"    \\%s at position %d\n",
					string(value[pos]),
					pos,
				))
			}
			msg = builder.String()
		}
		finding := makeLinterFinding(str, msg)

		// Fix
		var bytes []byte
		index := 0
		for _, backslashPos := range problems {
			for ; index < backslashPos; index++ {
				bytes = append(bytes, value[index])
			}
			bytes = append(bytes, '\\')
		}
		for ; index < len(value); index++ {
			bytes = append(bytes, value[index])
		}

		token := quotes + string(bytes) + quotes
		val, _, err := build.Unquote(token)
		if err == nil {
			newStr := *str
			newStr.Token = token
			newStr.Value = val
			finding.Replacement = []LinterReplacement{{expr, &newStr}}
		}

		findings = append(findings, finding)
	})
	return findings
}
