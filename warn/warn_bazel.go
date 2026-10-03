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

// General Bazel-related warnings

package warn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bazel-contrib/buildtools/v10/build"
)

// locationMakeVariableRe matches deprecated $(location) and $(locations) make variables.
var locationMakeVariableRe = regexp.MustCompile(`\$\(locations?(?:\s[^)]*)?\)`)

func constantGlobPatternWarning(patterns *build.ListExpr) []*LinterFinding {
	findings := []*LinterFinding{}
	for _, expr := range patterns.List {
		str, ok := expr.(*build.StringExpr)
		if !ok {
			continue
		}
		if !strings.Contains(str.Value, "*") {
			message := fmt.Sprintf(
				`Glob pattern %q has no wildcard ('*'). Constant patterns can be error-prone, move the file outside the glob.`, str.Value)
			findings = append(findings, makeLinterFinding(expr, message))
			return findings // at most one warning per glob
		}
	}
	return findings
}

func constantGlobWarning(f *build.File) []*LinterFinding {
	switch f.Type {
	case build.TypeBuild, build.TypeWorkspace, build.TypeBzl:
	default:
		// Not applicable
		return nil
	}

	findings := []*LinterFinding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		call, ok := expr.(*build.CallExpr)
		if !ok || len(call.List) == 0 {
			return
		}
		ident, ok := (call.X).(*build.Ident)
		if !ok || ident.Name != "glob" {
			return
		}
		patterns, ok := call.List[0].(*build.ListExpr)
		if ok {
			// first arg is unnamed and is a list
			findings = append(findings, constantGlobPatternWarning(patterns)...)
			return // at most one warning per glob
		}

		// look for named args called include
		for _, arg := range call.List {
			assignExpr, ok := arg.(*build.AssignExpr)
			if !ok {
				continue
			}
			str, ok := assignExpr.LHS.(*build.Ident)
			if !ok || str.Name != "include" {
				continue
			}
			patterns, ok := assignExpr.RHS.(*build.ListExpr)
			if ok {
				findings = append(findings, constantGlobPatternWarning(patterns)...)
				return // at most one warning per glob
			}
		}
	})
	return findings
}

func nativeInBuildFilesWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBuild {
		return nil
	}

	findings := []*LinterFinding{}
	build.WalkPointers(f, func(expr *build.Expr, stack []build.Expr) {
		// Search for `native.xxx` nodes
		dot, ok := (*expr).(*build.DotExpr)
		if !ok {
			return
		}
		ident, ok := dot.X.(*build.Ident)
		if !ok || ident.Name != "native" {
			return
		}

		findings = append(findings,
			makeLinterFinding(ident,
				`The "native" module shouldn't be used in BUILD files, its members are available as global symbols.`,
				LinterReplacement{expr, &build.Ident{Name: dot.Name}}))
	})
	return findings
}

func nativePackageWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBzl {
		return nil
	}

	findings := []*LinterFinding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Search for `native.package()` nodes
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := call.X.(*build.DotExpr)
		if !ok || dot.Name != "package" {
			return
		}
		ident, ok := dot.X.(*build.Ident)
		if !ok || ident.Name != "native" {
			return
		}

		findings = append(findings,
			makeLinterFinding(call, `"native.package()" shouldn't be used in .bzl files.`))
	})
	return findings
}

func duplicatedNameWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBuild && f.Type != build.TypeWorkspace {
		// Not applicable to .bzl files.
		return nil
	}

	findings := []*LinterFinding{}
	names := make(map[string]int) // map from name to line number
	msg := `A rule with name %q was already found on line %d. ` +
		`Even if it's valid for Blaze, this may confuse other tools. ` +
		`Please rename it and use different names.`

	for _, rule := range f.Rules("") {
		name := rule.ExplicitName()
		if name == "" {
			continue
		}
		start, _ := rule.Call.Span()
		if line, ok := names[name]; ok {
			finding := makeLinterFinding(rule.Call, fmt.Sprintf(msg, name, line))
			if nameNode := rule.Attr("name"); nameNode != nil {
				finding.Start, finding.End = nameNode.Span()
				start = finding.Start
			}
			findings = append(findings, finding)
		} else {
			names[name] = start.Line
		}
	}
	return findings
}

func positionalArgumentsWarning(f *build.File, fileReader *FileReader) (findings []*LinterFinding) {
	if f.Type != build.TypeBuild {
		return nil
	}
	macroAnalyzer := newMacroAnalyzer(fileReader)
	macroAnalyzer.files[f.Pkg+":"+f.Label] = analyzeFile(f)

	for _, expr := range f.Stmt {
		build.Walk(expr, func(x build.Expr, _ []build.Expr) {
			if fnCall, ok := x.(*build.CallExpr); ok {
				fnIdent, ok := fnCall.X.(*build.Ident)
				if !ok {
					return
				}

				if macroAnalyzer.IsRuleOrMacro(function{pkg: f.Pkg, filename: f.Label, name: fnIdent.Name}).isRuleOrMacro {
					for _, arg := range fnCall.List {
						if _, ok := arg.(*build.AssignExpr); ok || arg == nil {
							continue
						}
						findings = append(findings, makeLinterFinding(fnCall, fmt.Sprintf(
							`All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax.
Found call to rule or macro %q with positional arguments.`,
							fnIdent.Name)))
						return
					}
				}
			}
		})
	}
	return
}

func argsKwargsInBuildFilesWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBuild {
		return nil
	}

	findings := []*LinterFinding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// Search for function call nodes
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		for _, param := range call.List {
			unary, ok := param.(*build.UnaryExpr)
			if !ok {
				continue
			}
			switch unary.Op {
			case "*":
				findings = append(findings,
					makeLinterFinding(param, `*args are not allowed in BUILD files.`))
			case "**":
				findings = append(findings,
					makeLinterFinding(param, `**kwargs are not allowed in BUILD files.`))
			}
		}
	})
	return findings
}

func printWarning(f *build.File) []*LinterFinding {
	if f.Type == build.TypeDefault {
		// Only applicable to Bazel files
		return nil
	}

	findings := []*LinterFinding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		ident, ok := (call.X).(*build.Ident)
		if !ok || ident.Name != "print" {
			return
		}
		findings = append(findings,
			makeLinterFinding(expr, `"print()" is a debug function and shouldn't be submitted.`))
	})
	return findings
}

func makeLocationVariableWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBuild {
		return nil
	}

	findings := []*LinterFinding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		stringExpr, ok := expr.(*build.StringExpr)
		if !ok {
			return
		}
		if locationMakeVariableRe.MatchString(stringExpr.Value) {
			findings = append(findings,
				makeLinterFinding(stringExpr,
					`The "$(location)" and "$(locations)" make variables are deprecated. Use "$(execpath ...)" or "$(rootpath ...)" instead.`))
		}
	})
	return findings
}

// pathStringAttrs lists attributes that return a path string which may contain a configuration
// prefix (such as `bazel-out/k8-fastbuild/bin`) and is therefore not rewritten by path mapping:
// `File.path`, `File.dirname`, `File.root.path`, `ctx.bin_dir.path` and `ctx.genfiles_dir.path`.
var pathStringAttrs = map[string]bool{
	"path":    true,
	"dirname": true,
}

// isCtxActionsArgs reports whether expr is a `ctx.actions.args()` object. Besides the detected
// types, a variable named `args` is assumed to be one, e.g. a parameter of a helper function.
func isCtxActionsArgs(expr build.Expr, types map[build.Expr]Type) bool {
	if types[expr] == CtxActionsArgs {
		return true
	}
	ident, ok := expr.(*build.Ident)
	return ok && ident.Name == "args"
}

// pathMappingWarning reports path strings used in action command lines. Path mapping
// (https://github.com/bazelbuild/bazel/discussions/22658) can only rewrite `File` objects passed to
// `ctx.actions.args()`, so strings obtained from `File.path`, `File.dirname` etc. break actions
// that opt into it.
func pathMappingWarning(f *build.File) []*LinterFinding {
	if f.Type != build.TypeBzl {
		return nil
	}

	types := DetectTypes(f)
	findings := []*LinterFinding{}

	// checkArgument reports every path string expression inside a command line argument.
	checkArgument := func(arg build.Expr) {
		build.Walk(arg, func(expr build.Expr, stack []build.Expr) {
			dot, ok := expr.(*build.DotExpr)
			if !ok || !pathStringAttrs[dot.Name] {
				return
			}
			if len(stack) > 0 {
				// Skip method calls such as `ctx.path(...)` or `paths.dirname(...)`
				if call, ok := stack[len(stack)-1].(*build.CallExpr); ok && call.X == expr {
					return
				}
			}
			findings = append(findings, makeLinterFinding(dot, fmt.Sprintf(
				`%q is a path string that is not rewritten by path mapping. Pass the "File" object to "ctx.actions.args()" or compute the path in a "map_each" callback instead.`,
				build.FormatString(dot))))
		})
	}

	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		dot, ok := call.X.(*build.DotExpr)
		if !ok {
			return
		}

		switch {
		case isCtxActionsArgs(dot.X, types):
			// `args.add(...)`, `args.add_all(...)`, `args.add_joined(...)`
			switch dot.Name {
			case "add", "add_all", "add_joined":
			default:
				return
			}
			for _, arg := range call.List {
				if assign, ok := arg.(*build.AssignExpr); ok {
					arg = assign.RHS
				}
				checkArgument(arg)
			}
		case types[dot.X] == CtxActions:
			// `ctx.actions.run(...)` and `ctx.actions.run_shell(...)`
			var params []string
			switch dot.Name {
			case "run":
				params = []string{"executable", "arguments"}
			case "run_shell":
				params = []string{"command", "arguments"}
			default:
				return
			}
			for _, name := range params {
				if _, _, param := getParam(call.List, name); param != nil {
					checkArgument(param.RHS)
				}
			}
		}
	})
	return findings
}
