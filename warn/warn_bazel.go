package warn

import (
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
)

// Bazel-specific warnings

// negateExpression returns an expression which is a negation of the input.
// If it's a boolean literal (true or false), just return the opposite literal.
// If it's a unary expression with a unary `not` operator, just remove it.
// Otherwise, insert a `not` operator.
// It's assumed that input is no longer needed as it may be mutated or reused by the function.
func negateExpression(expr build.Expr) build.Expr {
	paren, ok := expr.(*build.ParenExpr)
	if ok {
		paren.X = negateExpression(paren.X)
		return paren
	}

	unary, ok := expr.(*build.UnaryExpr)
	if ok && unary.Op == "not" {
		return unary.X
	}

	boolean, ok := expr.(*build.Ident)
	if ok {
		if boolean.Name == "True" {
			boolean.Name = "False"
			return boolean
		} else if boolean.Name == "False" {
			boolean.Name = "True"
			return boolean
		}
	}

	return &build.UnaryExpr{
		Op: "not",
		X:  expr,
	}
}

// getParam search for a param with a given name in a given list of function arguments
// and returns it with its index
func getParam(attrs []build.Expr, paramName string) (int, *build.Ident, *build.BinaryExpr) {
	for i, attr := range attrs {
		binary, ok := attr.(*build.BinaryExpr)
		if !ok || binary.Op != "=" {
			continue
		}
		name, ok := (binary.X).(*build.Ident)
		if !ok || name.Name != paramName {
			continue
		}
		return i, name, binary
	}
	return -1, nil, nil
}

func getNewActionName(name string, params []build.Expr) (string, bool) {
	switch name {
	case "new_file":
		return "actions.declare_file", true
	case "experimental_new_directory":
		return "actions.declare_directory", true
	case "file_action":
		return "actions.write", true
	case "action":
		_, _, command := getParam(params, "command")
		if command != nil {
			return "actions.run_shell", true
		}
		return "actions.run", true
	case "empty_action":
		return "actions.do_nothing", true
	case "template_action":
		return "actions.expand_template", true
	default:
		return "", false
	}
}

// globalVariableUsageCheck checks whether there's a usage of a given global variable in the file.
// It's ok to shadow the name with a local variable and use it.
func globalVariableUsageCheck(f *build.File, category, global, alternative string, fix bool) []*Finding {
	findings := []*Finding{}

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		if ident, ok := (*e).(*build.Ident); ok {
			if ident.Name == global {
				if binding := env.Get(ident.Name); binding == nil {
					if fix {
						// It may be not correct to just replace the ident's name with `alternative` as it may be something complex
						// like `native.package_name()` which is not a valid ident, but it's fine for reformatting.
						ident.Name = alternative
					} else {
						start, end := ident.Span()
						findings = append(findings,
							makeFinding(f, start, end, category,
								"Global variable \""+global+"\" is deprecated in favor of \""+alternative+"\". Please rename it.", true, nil))
					}
				}
			}
		}
		bzlenv.WalkOnceWithEnvironment(*e, env, walk)
	}
	var expr build.Expr = f
	walk(&expr, bzlenv.NewEnvironment())

	return findings
}

func attrConfigurationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: attr.xxxx(..., cfg = "data", ...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "attr" {
			return
		}
		i, _, param := getParam(call.List, "cfg")
		if param == nil {
			return
		}
		value, ok := (param.Y).(*build.StringExpr)
		if !ok || value.Value != "data" {
			return
		}
		if fix {
			call.List = append(call.List[:i], call.List[i+1:]...)
		} else {
			start, end := param.Span()
			findings = append(findings,
				makeFinding(f, start, end, "attr-cfg",
					"cfg = \"data\" for attr definitions has no effect and should be removed.", true, nil))
		}
	})
	return findings
}

func attrNonEmptyWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: attr.xxxx(..., non_empty = ..., ...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "attr" {
			return
		}
		_, name, param := getParam(call.List, "non_empty")
		if param == nil {
			return
		}
		if fix {
			name.Name = "allow_empty"
			param.Y = negateExpression(param.Y)
		} else {
			start, end := param.Span()
			findings = append(findings,
				makeFinding(f, start, end, "attr-non-empty",
					"non_empty attributes for attr definitions are deprecated in favor of allow_empty.", true, nil))
		}
	})
	return findings
}

func attrSingleFileWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: attr.xxxx(..., single_file = ..., ...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "attr" {
			return
		}
		i, name, singleFileParam := getParam(call.List, "single_file")
		if singleFileParam == nil {
			return
		}
		if fix {
			value := singleFileParam.Y
			if boolean, ok := value.(*build.Ident); ok && boolean.Name == "False" {
				// if the value is `False`, just remove the whole parameter
				call.List = append(call.List[:i], call.List[i+1:]...)
			} else {
				// search for `allow_files` parameter in the same attr definition and remove it
				j, _, allowFilesParam := getParam(call.List, "allow_files")
				if allowFilesParam != nil {
					value = allowFilesParam.Y
					call.List = append(call.List[:j], call.List[j+1:]...)
				}
				singleFileParam.Y = value
				name.Name = "allow_single_file"
			}
		} else {
			start, end := singleFileParam.Span()
			findings = append(findings,
				makeFinding(f, start, end, "attr-single-file",
					"single_file is deprecated in favor of allow_single_file.", true, nil))
		}
	})
	return findings
}

func ctxActionsWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: ctx.xxxx(...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "ctx" {
			return
		}
		newName, ok := getNewActionName(dot.Name, call.List)
		if !ok {
			return
		}
		if fix {
			// Not entirely coorect because `newName` may be a sequence of attributes, e.g. `actions.run`.
			// But that's fine for formatting purposes.
			dot.Name = newName
		} else {
			start, end := dot.Span()
			findings = append(findings,
				makeFinding(f, start, end, "ctx-actions",
					"\"ctx."+dot.Name+"\" is deprecated in favor of \"ctx."+newName+"\".", true, nil))
		}
	})
	return findings
}

func fileTypeWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		defer bzlenv.WalkOnceWithEnvironment(*e, env, walk)

		call, ok := (*e).(*build.CallExpr)
		if !ok {
			return
		}
		ident, ok := (call.X).(*build.Ident)
		if !ok || ident.Name != "FileType" {
			return
		}
		if binding := env.Get("FileType"); binding == nil {
			start, end := call.Span()
			findings = append(findings,
				makeFinding(f, start, end, "filetype",
					"The FileType function is deprecated.", true, nil))
		}
	}
	var expr build.Expr = f
	walk(&expr, bzlenv.NewEnvironment())

	return findings
}

func packageNameWarning(f *build.File, fix bool) []*Finding {
	return globalVariableUsageCheck(f, "package-name", "PACKAGE_NAME", "native.package_name()", fix)
}

func repositoryNameWarning(f *build.File, fix bool) []*Finding {
	return globalVariableUsageCheck(f, "repository-name", "REPOSITORY_NAME", "native.repository_name()", fix)
}

func outputGroupWarning(f* build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Edit(f, func(expr build.Expr, stack []build.Expr) build.Expr {
		// Find nodes that match the following pattern: ctx.attr.xxx.output_group
		outputGroup, ok := (expr).(*build.DotExpr)
		if !ok || outputGroup.Name != "output_group"{
			return nil
		}
		dep, ok := (outputGroup.X).(*build.DotExpr)
		if !ok {
			return nil
		}
		attr, ok := (dep.X).(*build.DotExpr)
		if !ok || attr.Name != "attr" {
			return nil
		}
		ctx, ok := (attr.X).(*build.Ident)
		if !ok || ctx.Name != "ctx" {
			return nil
		}
		if !fix {
			start, end := outputGroup.Span()
			findings = append(findings,
				makeFinding(f, start, end, "output-group",
					"\"ctx.attr.dep.output_group\" is deprecated in favor of \"ctx.attr.dep[OutputGroupInfo]\".", true, nil))
			return nil
		}
		// Replace `xxx.output_group` with `xxx[OutputGroupInfo]`
		return &build.IndexExpr{
			X: dep,
			Y: &build.Ident{Name: "OutputGroupInfo"},
		}
	})
	return findings
}