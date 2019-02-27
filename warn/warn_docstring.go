package warn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

// FunctionLengthDocstringThreshold is a limit for a function size (in statements), above which
// a public function is required to have a docstring.
const FunctionLengthDocstringThreshold = 5

// getDocstrings returns a docstring of the statemenets and true if it exists.
// Otherwise it returns the first non-comment statement and false.
func getDocstring(stmts []build.Expr) (build.Expr, bool) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *build.CommentBlock:
			continue
		case *build.StringExpr:
			return stmt, true
		default:
			return stmt, false
		}
	}
	return nil, false
}

func moduleDocstringWarning(f *build.File, fix bool) []*Finding {
	if f.Type != build.TypeDefault && f.Type != build.TypeBzl {
		return []*Finding{}
	}
	if stmt, ok := getDocstring(f.Stmt); stmt != nil && !ok {
		start, end := stmt.Span()
		return []*Finding{
			makeFinding(f, start, end, "module-docstring",
				`The file has no module docstring.`, true, nil),
		}
	}
	return []*Finding{}
}

func stmtsCount(stmts []build.Expr) int {
	result := 0
	for _, stmt := range stmts {
		result++
		switch stmt := stmt.(type) {
		case *build.IfStmt:
			result += stmtsCount(stmt.True)
			result += stmtsCount(stmt.False)
		case *build.ForStmt:
			result += stmtsCount(stmt.Body)
		}
	}
	return result
}

// docstringInfo contains information about a function docstring
type docstringInfo struct {
	hasHeader    bool                      // whether the docstring has a one-line header
	args         map[string]build.Position // map of documented arguments, the values are line numbers
	returns      bool                      // whether the return value is documented
	argumentsPos build.Position            // line of the `Arguments:` block (not `Args:`), if it exists
}

// docstringBlock contains a block of a docstring (separated by empty lines)
type docstringBlock struct {
	startLineNo int      // line number of the first line of the block
	lines       []string // lines
}

// countLeadingSpaces returns the number of leading spaces of a string.
func countLeadingSpaces(s string) int {
	spaces := 0
	for _, c := range s {
		if c == ' ' {
			spaces++
		} else {
			break
		}
	}
	return spaces
}

var argRegex = regexp.MustCompile(`^ *(\*?\*?\w+)( *\([\w\ ,]+\))?:`)

// parseFunctionDocstring parses a function docstring and returns a docstringInfo object containing
// the parsed information about the function, its arguments and its return value.
func parseFunctionDocstring(doc *build.StringExpr) docstringInfo {
	start, _ := doc.Span()
	indent := start.LineRune - 1
	prefix := strings.Repeat(" ", indent)
	lines := strings.Split(doc.Value, "\n")

	info := docstringInfo{}
	info.args = make(map[string]build.Position)

	isArgumentsDescription := false // Whether the currently parsed block is an 'Args:' section
	argIndentation := 1000000       // Indentation at which previous arg documentation started

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}

	// The first non-empty line should be a single-line header
	for i, line := range lines {
		if line == "" {
			continue
		}
		if i == len(lines)-1 || lines[i+1] == "" {
			info.hasHeader = true
		}
		break
	}

	// Search for Args: and Returns: sections
	for i, line := range lines {
		switch line {
		case prefix + "Arguments:":
			info.argumentsPos = build.Position{
				Line:     start.Line + i,
				LineRune: indent,
			}
			isArgumentsDescription = true
			continue
		case prefix + "Args:":
			isArgumentsDescription = true
			continue
		case prefix + "Returns:":
			info.returns = true
			continue
		}

		if isArgumentsDescription {
			newIndentation := countLeadingSpaces(line)

			if line != "" && newIndentation <= indent {
				// The indented block is over
				isArgumentsDescription = false
				continue
			} else if newIndentation > argIndentation {
				// Continuation of the previous argument description
				continue
			} else {
				// Maybe a new argument is described here
				result := argRegex.FindStringSubmatch(line)
				if len(result) > 1 {
					argIndentation = newIndentation
					info.args[result[1]] = build.Position{
						Line:     start.Line + i,
						LineRune: indent + argIndentation,
					}
				}
			}
		}
	}
	return info
}

func getParamName(param build.Expr) string {
	switch param := param.(type) {
	case *build.Ident:
		return param.Name
	case *build.AssignExpr:
		// keyword parameter
		if ident, ok := param.LHS.(*build.Ident); ok {
			return ident.Name
		}
	case *build.UnaryExpr:
		// *args or **kwargs
		if ident, ok := param.X.(*build.Ident); ok {
			return param.Op + ident.Name
		}
	}
	return ""
}

func hasReturnValues(def *build.DefStmt) bool {
	result := false
	build.Walk(def, func(expr build.Expr, stack []build.Expr) {
		ret, ok := expr.(*build.ReturnStmt)
		if ok && ret.Result != nil {
			result = true
		}
	})
	return result
}

func functionDocstringWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	for _, stmt := range f.Stmt {
		def, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		// A docstring is required for public functions if they are long enough (at least 5 statements)
		isDocstringRequired := !strings.HasPrefix(def.Name, "_") && stmtsCount(def.Body) >= FunctionLengthDocstringThreshold

		doc, ok := getDocstring(def.Body)
		if !ok {
			if isDocstringRequired {
				// Public functions that are not too short should have a docstring
				start, end := stmt.Span()
				findings = append(findings, makeFinding(f, start, end, "function-docstring",
					fmt.Sprintf(`The function "%s" has no docstring.`, def.Name), true, nil))
			}
			continue
		}

		// Docstring exists, check for its contents
		info := parseFunctionDocstring(doc.(*build.StringExpr))
		start, end := doc.Span()

		if !info.hasHeader {
			findings = append(findings, makeFinding(f, start, end, "function-docstring",
				fmt.Sprintf(`The docstring for the function "%s" should start with a one-line summary.`, def.Name), true, nil))
		}
		if info.argumentsPos.LineRune > 0 {
			argumentsEnd := info.argumentsPos
			argumentsEnd.LineRune += len("Arguments:")
			findings = append(findings, makeFinding(f, info.argumentsPos, argumentsEnd, "function-docstring",
				`Prefer 'Args:' to 'Arguments:' when documenting function arguments.`, true, nil))
		}

		// If the docstring is required or there are any arguments described, check for their integrity.
		if isDocstringRequired || len(info.args) > 0 {

			// Check whether all arguments are documented.
			notDocumentedArguments := []string{}
			paramNames := make(map[string]bool)
			for _, param := range def.Params {
				name := getParamName(param)
				paramNames[name] = true
				if _, ok := info.args[name]; !ok {
					notDocumentedArguments = append(notDocumentedArguments, name)
				}
			}
			if len(notDocumentedArguments) > 0 {
				if len(notDocumentedArguments) == 1 {
					findings = append(findings, makeFinding(f, start, end, "function-docstring",
						fmt.Sprintf(`Argument "%s" is not documented.`, notDocumentedArguments[0]), true, nil))
				} else {
					findings = append(findings, makeFinding(f, start, end, "function-docstring",
						fmt.Sprintf(
							`Arguments "%s" are not documented.`,
							strings.Join(notDocumentedArguments, `", "`),
						), true, nil))
				}
			}

			// Check whether all documented arguments actually exist in the function signature.
			for name, pos := range info.args {
				if !paramNames[name] {
					posEnd := pos
					posEnd.LineRune += len(name)
					findings = append(findings, makeFinding(f, pos, posEnd, "function-docstring",
						fmt.Sprintf(`Argument "%s" is documented but doesn't exist in the function signature.`, name), true, nil))
				}
			}
		}

		// Check whether the return value is documented
		if isDocstringRequired && hasReturnValues(def) && !info.returns {
			findings = append(findings, makeFinding(f, start, end, "function-docstring",
				fmt.Sprintf(`Return value of "%s" is not documented.`, def.Name), true, nil))
		}
	}
	return findings
}
