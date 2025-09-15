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

// Warnings for using deprecated functions

package warn

import (
	"fmt"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/labels"
)

// function represents a function identifier, which is a pair (module name, function name).
type function struct {
	pkg      string // package where the function is defined
	filename string // name of a .bzl file relative to the package
	name     string // original name of the function
}

func (f function) label() string {
	return f.pkg + ":" + f.filename
}

// callStackFrame formats a function call to be printable in a call stack.
func (f function) callStackFrame(ref build.Expr) string {
	return fmt.Sprintf("%s:%d: %s", f.label(), exprLine(ref), f.name)
}

// acceptsNameArgument checks whether a function can accept a named argument called "name",
// either directly or via **kwargs.
func acceptsNameArgument(def *build.DefStmt) bool {
	for _, param := range def.Params {
		if name, op := build.GetParamName(param); name == "name" || op == "**" {
			return true
		}
	}
	return false
}

// exprLine returns the start line of an expression
func exprLine(expr build.Expr) int {
	start, _ := expr.Span()
	return start.Line
}

// functionReport represents the analysis result of a function
type macroReport struct {
	callStack []string
}

// isMacroOrRule returns true if the report contained a rule or a macro.
func (mr macroReport) isMacroOrRule() bool {
	return mr.callStack != nil
}

// wrapReport is a helper to wrap a marcoReport for a downstream call.
func wrapReport(fn function, ref build.Expr, mr macroReport) macroReport {
	return macroReport{callStack: append([]string{fn.callStackFrame(ref)}, mr.callStack...)}
}

// macroAnalyzer is an object that analyzes the directed graph of functions calling each other,
// loading other files lazily if necessary.
type macroAnalyzer struct {
	fileReader *FileReader
	// Local files if file is defined here, fileReader is not used.
	localFiles map[string]*build.File
	cache      map[function]*macroReport
}

func (ma macroAnalyzer) getFile(fn function) *build.File {
	f := ma.localFiles[fn.label()]
	if f != nil {
		return f
	}
	if ma.fileReader == nil {
		return nil
	}
	return ma.fileReader.GetFile(fn.pkg, fn.filename)
}

// AnalyzeFn is a public function that checks whether the given function is a macro or rule.
func (ma macroAnalyzer) AnalyzeFn(fn function) (report macroReport) {
	// Check the cache first
	if cached, ok := ma.cache[fn]; ok {
		return *cached
	}
	// Write an empty result to the cache before analyzing. This will prevent stack overflow crashes
	// if the input data contains recursion.
	ma.cache[fn] = &macroReport{}
	defer func() {
		// Update the cache with the actual result
		ma.cache[fn] = &report
	}()

	f := ma.getFile(fn)
	if f == nil {
		return macroReport{}
	}

	for _, stmt := range f.Stmt {
		switch stmt := stmt.(type) {
		// If function is loaded from another file, check the separate file for the function call.
		case *build.LoadStmt:
			for i, from := range stmt.From {
				if stmt.To[i].Name == fn.name {
					label := labels.ParseRelative(stmt.Module.Value, f.Pkg)
					if r := ma.AnalyzeFn(function{label.Package, label.Target, from.Name}); r.isMacroOrRule() {
						return r
					}
				}
			}
		case *build.AssignExpr:
			// Analyze aliases (`foo = bar`) or rule declarations (`foo = rule(...)`)
			if lhsIdent, ok := stmt.LHS.(*build.Ident); !ok || lhsIdent.Name != fn.name {
				continue
			}

			// If the RHS is an identifier (LHS is an alias), check if RHS is a macro.
			if rhsIdent, ok := stmt.RHS.(*build.Ident); ok {
				if r := ma.AnalyzeFn(function{f.Pkg, f.Label, rhsIdent.Name}); r.isMacroOrRule() {
					return wrapReport(fn, stmt, r)
				}
				continue
			}

			// If the RHS is a function call, check if the called function is a "rule" or a macro.
			if call, ok := stmt.RHS.(*build.CallExpr); ok {
				if ident, ok := call.X.(*build.Ident); ok {
					if ident.Name == "macro" {
						report = macroReport{callStack: []string{
							function{f.Pkg, f.Label, "(MACRO)"}.callStackFrame(stmt),
						}}
						return
					}
					if ident.Name == "rule" {
						report = macroReport{callStack: []string{
							function{f.Pkg, f.Label, "(RULE)"}.callStackFrame(stmt),
						}}
						return
					}
					if r := ma.AnalyzeFn(function{f.Pkg, f.Label, ident.Name}); r.isMacroOrRule() {
						return wrapReport(fn, stmt, r)
					}
				}
			}

			if dotExpr, ok := stmt.RHS.(*build.DotExpr); ok {
				// Note: Currently only handles "native." dot-expressions, others are ignored.
				if isNativeRule(dotExpr) {
					return macroReport{callStack: []string{function{f.Pkg, f.Label, "native." + dotExpr.Name + " (NATIVE RULE)"}.callStackFrame(dotExpr)}}
				}
			}
		case *build.DefStmt:
			if stmt.Name != fn.name {
				continue
			}
			// If the function is implemented here, check if it calls any rules or macros.
			if r := ma.checkFunctionCalls(fn, stmt); r.isMacroOrRule() {
				return wrapReport(fn, stmt, r)
			}
		default:
			continue
		}
	}
	return
}

func (ma macroAnalyzer) checkFunctionCalls(fn function, def *build.DefStmt) (report macroReport) {
	report = macroReport{}
	build.Walk(def, func(expr build.Expr, stack []build.Expr) {
		if call, ok := expr.(*build.CallExpr); ok {
			if fnIdent, ok := call.X.(*build.Ident); ok {
				calledFn := function{pkg: fn.pkg, filename: fn.filename, name: fnIdent.Name}
				if r := ma.AnalyzeFn(calledFn); r.isMacroOrRule() {
					report = wrapReport(calledFn, call, r)
					return
				}
			}
		}
		if dotExpr, ok := expr.(*build.DotExpr); ok {
			// Note: Currently only handles "native." dot-expressions, others are ignored.
			if isNativeRule(dotExpr) {
				dotFn := function{pkg: fn.pkg, filename: fn.filename, name: "native." + dotExpr.Name + " (NATIVE RULE)"}
				report = macroReport{callStack: []string{dotFn.callStackFrame(dotExpr)}}
			}
		}
	})
	return report
}

func isNativeRule(expr *build.DotExpr) bool {
	if ident, ok := expr.X.(*build.Ident); !ok || ident.Name != "native" {
		return false
	}

	switch expr.Name {
	case "glob", "existing_rule", "existing_rules", "package_name",
		"repository_name", "exports_files":

		// Not a rule
		return false
	default:
		return true
	}
}

// newMacroAnalyzer creates and initiates an instance of macroAnalyzer.
func newMacroAnalyzer(fileReader *FileReader) macroAnalyzer {
	return macroAnalyzer{
		fileReader: fileReader,
		localFiles: make(map[string]*build.File),
		cache:      make(map[function]*macroReport),
	}
}

func unnamedMacroWarning(f *build.File, fileReader *FileReader) []*LinterFinding {
	if f.Type != build.TypeBzl {
		return nil
	}

	macroAnalyzer := newMacroAnalyzer(fileReader)
	macroAnalyzer.localFiles[f.Pkg+":"+f.Label] = f

	var findings []*LinterFinding
	for _, stmt := range f.Stmt {
		def, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		if strings.HasPrefix(def.Name, "_") || acceptsNameArgument(def) {
			continue
		}

		fn := function{f.Pkg, f.Label, def.Name}
		report := macroAnalyzer.AnalyzeFn(fn)
		if !report.isMacroOrRule() {
			continue
		}
		msg := fmt.Sprintf(`The macro %q should have a keyword argument called "name".

It is considered a macro because it calls a rule or another macro, call stack:

%s

By convention, every public macro needs a "name" argument (even if it doesn't use it).
This is important for tooling and automation.

  * If this function is a helper function that's not supposed to be used outside of this file,
    please make it private (e.g. rename it to "_%s").
  * Otherwise, add a "name" argument. If possible, use that name when calling other macros/rules.`,
			def.Name,
			strings.Join(report.callStack, "\n"),
			def.Name)
		finding := makeLinterFinding(def, msg)
		finding.End = def.ColonPos
		findings = append(findings, finding)
	}

	return findings
}
