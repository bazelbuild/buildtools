// Warnings for incompatible changes in the Bazel API

package warn

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
	"github.com/bazelbuild/buildtools/edit"
)

// Bazel API-specific warnings

var functionsWithPositionalArguments = map[string]bool{
	"distribs":            true,
	"exports_files":       true,
	"licenses":            true,
	"print":               true,
	"register_toolchains": true,
	"vardef":              true,
}

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

// isFunctionCall checks whether expr is a call of a function with a given name
func isFunctionCall(expr build.Expr, name string) (*build.CallExpr, bool) {
	call, ok := expr.(*build.CallExpr)
	if !ok {
		return nil, false
	}
	if ident, ok := call.X.(*build.Ident); ok && ident.Name == name {
		return call, true
	}
	return nil, false
}

// globalVariableUsageCheck checks whether there's a usage of a given global variable in the file.
// It's ok to shadow the name with a local variable and use it.
func globalVariableUsageCheck(f *build.File, category, global, alternative string, fix bool) []*Finding {
	findings := []*Finding{}

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		defer bzlenv.WalkOnceWithEnvironment(*e, env, walk)

		ident, ok := (*e).(*build.Ident)
		if !ok {
			return
		}
		if ident.Name != global {
			return
		}
		if binding := env.Get(ident.Name); binding != nil {
			return
		}
		if fix {
			// It may be not correct to just replace the ident's name with `alternative` as it may be something complex
			// like `native.package_name()` which is not a valid ident, but it's fine for reformatting.
			ident.Name = alternative
			return
		}
		start, end := ident.Span()
		findings = append(findings,
			makeFinding(f, start, end, category,
				fmt.Sprintf(`Global variable "%s" is deprecated in favor of "%s". Please rename it.`, global, alternative), true, nil))
	}
	var expr build.Expr = f
	walk(&expr, bzlenv.NewEnvironment())

	return findings
}

// notLoadedFunctionUsageCheck checks whether there's a usage of a given not imported  function in the file
// and adds a load statement if necessary.
func notLoadedFunctionUsageCheck(f *build.File, category string, globals []string, loadFrom string, fix bool) []*Finding {
	findings := []*Finding{}
	toLoad := []string{}

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		defer bzlenv.WalkOnceWithEnvironment(*e, env, walk)

		call, ok := (*e).(*build.CallExpr)
		if !ok {
			return
		}

		ident, ok := (call.X).(*build.Ident)
		if !ok {
			return
		}

		if binding := env.Get(ident.Name); binding != nil {
			return
		}
		for _, global := range globals {
			if ident.Name == global {
				if fix {
					toLoad = append(toLoad, global)
					return
				}
				start, end := call.Span()
				findings = append(findings,
					makeFinding(f, start, end, category,
						fmt.Sprintf(`Function "%s" is not global anymore and needs to be loaded from "%s".`, global, loadFrom), true, nil))
			}
		}
	}
	var expr build.Expr = f
	walk(&expr, bzlenv.NewEnvironment())

	if fix && len(toLoad) > 0 {
		f.Stmt = edit.InsertLoad(f.Stmt, loadFrom, toLoad, toLoad)
	}

	return findings
}

// makePositional makes the function argument positional (removes the keyword if it exists)
func makePositional(argument build.Expr) build.Expr {
	if binary, ok := argument.(*build.BinaryExpr); ok {
		return binary.Y
	}
	return argument
}

// makeKeyword makes the function argument keyword (adds or edits the keyword name)
func makeKeyword(argument build.Expr, name string) build.Expr {
	binary, ok := argument.(*build.BinaryExpr)
	if !ok {
		return &build.BinaryExpr{
			X:  &build.Ident{Name: name},
			Op: "=",
			Y:  argument,
		}
	}
	ident, ok := binary.X.(*build.Ident)
	if !ok {
		binary.X = &build.Ident{Name: name}
		return binary
	}
	ident.Name = name
	return argument
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
			return
		}
		start, end := param.Span()
		findings = append(findings,
			makeFinding(f, start, end, "attr-cfg",
				`cfg = "data" for attr definitions has no effect and should be removed.`, true, nil))
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
		if !fix {
			start, end := singleFileParam.Span()
			findings = append(findings,
				makeFinding(f, start, end, "attr-single-file",
					"single_file is deprecated in favor of allow_single_file.", true, nil))
			return
		}
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
	})
	return findings
}

func ctxActionsWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	addWarning := func(expr build.Expr, name string) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "ctx-actions",
				fmt.Sprintf(`"ctx.%s" is deprecated.`, name), true, nil))
	}

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
		if !fix {
			switch dot.Name {
			case "new_file", "experimental_new_directory", "file_action", "action", "empty_action", "template_action":
				addWarning(dot, dot.Name)
			}
			return
		}
		switch dot.Name {
		case "new_file":
			if len(call.List) > 2 {
				// Can't fix automatically because the new API doesn't support the 3 arguments signature
				addWarning(dot, dot.Name)
				return
			}
			dot.Name = "actions.declare_file"
			if len(call.List) == 2 {
				// swap arguments:
				// ctx.new_file(sibling, name) -> ctx.actions.declare_file(name, sibling=sibling)
				call.List[0], call.List[1] = makePositional(call.List[1]), makeKeyword(call.List[0], "sibling")
			}
		case "experimental_new_directory":
			dot.Name = "actions.declare_directory"
		case "file_action":
			dot.Name = "actions.write"
			_, ident, _ := getParam(call.List, "executable")
			if ident != nil {
				ident.Name = "is_executable"
			}
		case "action":
			dot.Name = "actions.run"
			_, _, command := getParam(call.List, "command")
			if command != nil {
				dot.Name = "actions.run_shell"
			}
		case "empty_action":
			dot.Name = "actions.do_nothing"
		case "template_action":
			dot.Name = "actions.expand_template"
			if _, ident, _ := getParam(call.List, "executable"); ident != nil {
				ident.Name = "is_executable"
			}
		}
		return
	})
	return findings
}

func fileTypeWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	var walk func(e *build.Expr, env *bzlenv.Environment)
	walk = func(e *build.Expr, env *bzlenv.Environment) {
		defer bzlenv.WalkOnceWithEnvironment(*e, env, walk)

		call, ok := isFunctionCall(*e, "FileType")
		if !ok {
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

func outputGroupWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Edit(f, func(expr build.Expr, stack []build.Expr) build.Expr {
		// Find nodes that match the following pattern: ctx.attr.xxx.output_group
		outputGroup, ok := (expr).(*build.DotExpr)
		if !ok || outputGroup.Name != "output_group" {
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
					`"ctx.attr.dep.output_group" is deprecated in favor of "ctx.attr.dep[OutputGroupInfo]".`, true, nil))
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

func nativeGitRepositoryWarning(f *build.File, fix bool) []*Finding {
	return notLoadedFunctionUsageCheck(f, "git-repository", []string{"git_repository", "new_git_repository"}, "@bazel_tools//tools/build_defs/repo:git.bzl", fix)
}

func nativeHTTPArchiveWarning(f *build.File, fix bool) []*Finding {
	return notLoadedFunctionUsageCheck(f, "http-archive", []string{"http_archive"}, "@bazel_tools//tools/build_defs/repo:http.bzl", fix)
}

func contextArgsAPIWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	types := detectTypes(f)

	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Search for `<ctx.actions.args>.add()` nodes
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := call.X.(*build.DotExpr)
		if !ok || dot.Name != "add" || types[dot.X] != CtxActionsArgs {
			return
		}

		// If neither before_each nor join_with nor map_fn is specified, the node is ok.
		// Otherwise if join_with is specified, use `.add_joined` instead.
		// Otherwise use `.add_all` instead.

		_, beforeEachKw, beforeEach := getParam(call.List, "before_each")
		_, _, joinWith := getParam(call.List, "join_with")
		_, mapFnKw, mapFn := getParam(call.List, "map_fn")
		if beforeEach == nil && joinWith == nil && mapFn == nil {
			// No deprecated API detected
			return
		}
		if !fix {
			start, end := call.Span()
			findings = append(findings,
				makeFinding(f, start, end, "ctx-args",
					`"ctx.actions.args().add()" for multiple arguments is deprecated in favor of "add_all()" or "add_joined()".`, true, nil))
			return
		}

		dot.Name = "add_all"
		if joinWith != nil {
			dot.Name = "add_joined"
			if beforeEach != nil {
				// `add_joined` doesn't have a `before_each` parameter, replace it with `format_each`:
				// `before_each = foo` -> `format_each = foo + "%s"`
				beforeEachKw.Name = "format_each"
				beforeEach.Y = &build.BinaryExpr{
					X:  beforeEach.Y,
					Op: "+",
					Y:  &build.StringExpr{Value: "%s"},
				}
			}
		}
		if mapFnKw != nil {
			mapFnKw.Name = "map_each"
		}
	})
	return findings
}

func attrOutputDefaultWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: attr.output(..., default = ...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok || dot.Name != "output" {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "attr" {
			return
		}
		_, _, param := getParam(call.List, "default")
		if param == nil {
			return
		}
		start, end := param.Span()
		findings = append(findings,
			makeFinding(f, start, end, "attr-output-default",
				`The "default" parameter for attr.output() is deprecated.`, true, nil))
	})
	return findings
}

func attrLicenseWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Find nodes that match the following pattern: attr.license(...)
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := (call.X).(*build.DotExpr)
		if !ok || dot.Name != "license" {
			return
		}
		base, ok := dot.X.(*build.Ident)
		if !ok || base.Name != "attr" {
			return
		}
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "attr-license",
				`"attr.license()" is deprecated and shouldn't be used.`, true, nil))
	})
	return findings
}

// ruleImplReturnWarning checks whether a rule implementation function returns an old-style struct
func ruleImplReturnWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	// iterate over rules and collect rule implementation function names
	implNames := make(map[string]bool)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		call, ok := isFunctionCall(expr, "rule")
		if !ok {
			return
		}

		// Try to get the implementaton parameter either by name or as the first argument
		var impl build.Expr
		_, _, param := getParam(call.List, "implementation")
		if param != nil {
			impl = param.Y
		} else if len(call.List) > 0 {
			impl = call.List[0]
		}
		if name, ok := impl.(*build.Ident); ok {
			implNames[name.Name] = true
		}
	})

	// iterate over functions
	for _, stmt := range f.Stmt {
		def, ok := stmt.(*build.DefStmt)
		if !ok || !implNames[def.Name] {
			// either not a function or not used in the file as a rule implementation function
			continue
		}
		// traverse the function and find all of its return statements
		build.Walk(def, func(expr build.Expr, stack []build.Expr) {
			ret, ok := expr.(*build.ReturnStmt)
			if !ok {
				return
			}
			// check whether it returns a struct
			if _, ok := isFunctionCall(ret.Result, "struct"); ok {
				start, end := ret.Span()
				findings = append(findings,
					makeFinding(f, start, end, "rule-impl-return",
						`Avoid using the legacy provider syntax.`, true, nil))
			}
		})
	}

	return findings
}
