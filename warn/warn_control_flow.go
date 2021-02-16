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

// Warnings related to the control flow

package warn

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/bzlenv"
	"github.com/bazelbuild/buildtools/edit"
)

// findReturnsWithoutValue searches for return statements without a value, calls `callback` on
// them and returns whether the current list of statements terminates (either by a return or fail()
// statements on the current level in all subranches.
func findReturnsWithoutValue(stmts []build.Expr, callback func(*build.ReturnStmt)) bool {
	if len(stmts) == 0 {
		// May occur in empty else-clauses
		return false
	}
	terminated := false
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *build.ReturnStmt:
			if stmt.Result == nil {
				callback(stmt)
			}
			terminated = true
		case *build.CallExpr:
			ident, ok := stmt.X.(*build.Ident)
			if ok && ident.Name == "fail" {
				terminated = true
			}
		case *build.ForStmt:
			// Call recursively to find all return statements without a value there.
			// Even if a for-loop is guaranteed to terminate in each iteration, buildifier still can't
			// check whether the loop is not empty, so we can't say that the statement after the ForStmt
			// is unreachable.
			findReturnsWithoutValue(stmt.Body, callback)
		case *build.IfStmt:
			// Save to separate values to avoid short circuit evaluation
			term1 := findReturnsWithoutValue(stmt.True, callback)
			term2 := findReturnsWithoutValue(stmt.False, callback)
			if term1 && term2 {
				terminated = true
			}
		}
	}
	return terminated
}

// missingReturnValueWarning warns if a function returns both explicit and implicit values.
func missingReturnValueWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}

	for _, stmt := range f.Stmt {
		function, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		var hasNonEmptyReturns bool
		build.Walk(function, func(expr build.Expr, stack []build.Expr) {
			if ret, ok := expr.(*build.ReturnStmt); ok && ret.Result != nil {
				hasNonEmptyReturns = true
			}
		})

		if !hasNonEmptyReturns {
			continue
		}
		explicitReturn := findReturnsWithoutValue(function.Body, func(ret *build.ReturnStmt) {
			findings = append(findings,
				makeLinterFinding(ret, fmt.Sprintf("Some but not all execution paths of %q return a value.", function.Name)))
		})
		if !explicitReturn {
			findings = append(findings,
				makeLinterFinding(function, fmt.Sprintf(`Some but not all execution paths of %q return a value.
The function may terminate by an implicit return in the end.`, function.Name)))
		}
	}
	return findings
}

// findUnreachableStatements searches for unreachable statements (i.e. statements that immediately
// follow `return`, `break`, `continue`, and `fail()` statements and calls `callback` on them.
// If there are several consequent unreachable statements, it only reports the first of them.
// Returns whether the execution is terminated explicitly.
func findUnreachableStatements(stmts []build.Expr, callback func(build.Expr)) bool {
	unreachable := false
	for _, stmt := range stmts {
		if unreachable {
			callback(stmt)
			return true
		}
		switch stmt := stmt.(type) {
		case *build.ReturnStmt:
			unreachable = true
		case *build.CallExpr:
			ident, ok := stmt.X.(*build.Ident)
			if ok && ident.Name == "fail" {
				unreachable = true
			}
		case *build.BranchStmt:
			if stmt.Token != "pass" {
				// either break or continue
				unreachable = true
			}
		case *build.ForStmt:
			findUnreachableStatements(stmt.Body, callback)
		case *build.IfStmt:
			// Save to separate values to avoid short circuit evaluation
			term1 := findUnreachableStatements(stmt.True, callback)
			term2 := findUnreachableStatements(stmt.False, callback)
			if term1 && term2 {
				unreachable = true
			}
		}
	}
	return unreachable
}

func unreachableStatementWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}

	for _, stmt := range f.Stmt {
		function, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		findUnreachableStatements(function.Body, func(expr build.Expr) {
			findings = append(findings,
				makeLinterFinding(expr, `The statement is unreachable.`))
		})
	}
	return findings
}

func noEffectStatementsCheck(body []build.Expr, isTopLevel, isFunc bool, findings []*LinterFinding) []*LinterFinding {
	seenNonComment := false
	for _, stmt := range body {
		if stmt == nil {
			continue
		}

		_, isString := stmt.(*build.StringExpr)
		if isString {
			if !seenNonComment && (isTopLevel || isFunc) {
				// It's a docstring.
				seenNonComment = true
				continue
			}
		}
		if _, ok := stmt.(*build.CommentBlock); !ok {
			seenNonComment = true
		}
		switch s := (stmt).(type) {
		case *build.DefStmt, *build.ForStmt, *build.IfStmt, *build.LoadStmt, *build.ReturnStmt,
			*build.CallExpr, *build.CommentBlock, *build.BranchStmt, *build.AssignExpr:
			continue
		case *build.Comprehension:
			if !isTopLevel || s.Curly {
				// List comprehensions are allowed on top-level.
				findings = append(findings,
					makeLinterFinding(stmt, "Expression result is not used. Use a for-loop instead of a list comprehension."))
			}
			continue
		}

		msg := "Expression result is not used."
		if isString {
			msg += " Docstrings should be the first statements of a file or a function (they may follow comment lines)."
		}
		findings = append(findings, makeLinterFinding(stmt, msg))
	}
	return findings
}

func noEffectWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}
	findings = noEffectStatementsCheck(f.Stmt, true, false, findings)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// The AST should have a ExprStmt node.
		// Since we don't have that, we match on the nodes that contain a block to get the list of statements.
		switch expr := expr.(type) {
		case *build.ForStmt:
			findings = noEffectStatementsCheck(expr.Body, false, false, findings)
		case *build.DefStmt:
			findings = noEffectStatementsCheck(expr.Function.Body, false, true, findings)
		case *build.IfStmt:
			findings = noEffectStatementsCheck(expr.True, false, false, findings)
			findings = noEffectStatementsCheck(expr.False, false, false, findings)
		}
	})
	return findings
}

// unusedVariableCheck checks for unused variables inside a given node `stmt` (either *build.File or
// *build.DefStmt) and reports unused and already defined variables.
func unusedVariableCheck(f *build.File, stmts []build.Expr, findings []*LinterFinding) []*LinterFinding {
	if f.Type == build.TypeDefault || f.Type == build.TypeBzl {
		// Not applicable to .bzl files, unused symbols may be loaded and used in other files.
		return findings
	}
	usedSymbols := make(map[string]bool)

	for _, stmt := range stmts {
		for key := range edit.UsedSymbols(stmt) {
			usedSymbols[key] = true
		}
	}

	for _, s := range stmts {
		if defStmt, ok := s.(*build.DefStmt); ok {
			findings = unusedVariableCheck(f, defStmt.Body, findings)
			continue
		}

		// look for all assignments in the scope
		as, ok := s.(*build.AssignExpr)
		if !ok {
			continue
		}
		left, ok := as.LHS.(*build.Ident)
		if !ok {
			continue
		}
		if usedSymbols[left.Name] {
			continue
		}
		if edit.ContainsComments(s, "@unused") {
			// To disable the warning, put a comment that contains '@unused'
			continue
		}
		findings = append(findings,
			makeLinterFinding(as.LHS, fmt.Sprintf(`Variable %q is unused. Please remove it.
To disable the warning, add '@unused' in a comment.`, left.Name)))
	}
	return findings
}

func unusedVariableWarning(f *build.File) []*LinterFinding {
	return unusedVariableCheck(f, f.Stmt, []*LinterFinding{})
}

func redefinedVariableWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}
	definedSymbols := make(map[string]bool)

	types := detectTypes(f)
	for _, s := range f.Stmt {
		// look for all assignments in the scope
		as, ok := s.(*build.AssignExpr)
		if !ok {
			continue
		}
		left, ok := as.LHS.(*build.Ident)
		if !ok {
			continue
		}
		if !definedSymbols[left.Name] {
			definedSymbols[left.Name] = true
			continue
		}

		if as.Op == "+=" && (types[as.LHS] == List || types[as.RHS] == List) {
			// Not a reassignment, just appending to a list
			continue
		}

		findings = append(findings,
			makeLinterFinding(as.LHS, fmt.Sprintf(`Variable %q has already been defined. 
Redefining a global value is discouraged and will be forbidden in the future.
Consider using a new variable instead.`, left.Name)))
	}
	return findings
}

func unusedLoadWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}
	loaded := make(map[string]struct {
		label, from string
		line        int
	})

	symbols := edit.UsedSymbols(f)
	for stmtIndex := 0; stmtIndex < len(f.Stmt); stmtIndex++ {
		originalLoad, ok := f.Stmt[stmtIndex].(*build.LoadStmt)
		if !ok {
			continue
		}

		// Findings related to the current load statement
		loadFindings := []*LinterFinding{}

		// Copy the `load` object to provide a replacement if needed
		load := *originalLoad
		load.From = append([]*build.Ident{}, load.From...)
		load.To = append([]*build.Ident{}, load.To...)

		for i := 0; i < len(load.To); i++ {
			from := load.From[i]
			to := load.To[i]
			// Check if the symbol was already loaded
			origin, alreadyLoaded := loaded[to.Name]
			start, _ := from.Span()
			loaded[to.Name] = struct {
				label, from string
				line        int
			}{load.Module.Token, from.Name, start.Line}

			if alreadyLoaded {
				// The same symbol has already been loaded earlier
				if origin.label == load.Module.Token && origin.from == from.Name {
					// Only fix if it's loaded from the same label and variable
					load.To = append(load.To[:i], load.To[i+1:]...)
					load.From = append(load.From[:i], load.From[i+1:]...)
					i--
                                	loadFindings = append(loadFindings, makeLinterFinding(to,
                                        	fmt.Sprintf("Symbol %q has already been loaded on line %d. Please remove it.", to.Name, origin.line)))
	                                continue
				}

				loadFindings = append(loadFindings, makeLinterFinding(to,
					fmt.Sprintf("A different symbol %q has already been loaded on line %d. Please use a different local name.", to.Name, origin.line)))
				continue
			}
			_, ok := symbols[to.Name]
			if !ok && !edit.ContainsComments(originalLoad, "@unused") && !edit.ContainsComments(to, "@unused") && !edit.ContainsComments(from, "@unused") {
				// The loaded symbol is not used and is not protected by a special "@unused" comment
				load.To = append(load.To[:i], load.To[i+1:]...)
				load.From = append(load.From[:i], load.From[i+1:]...)
				i--

				loadFindings = append(loadFindings, makeLinterFinding(to,
					fmt.Sprintf("Loaded symbol %q is unused. Please remove it.\nTo disable the warning, add '@unused' in a comment.", to.Name)))
				if f.Type == build.TypeDefault || f.Type == build.TypeBzl {
					loadFindings[len(loadFindings)-1].Message += fmt.Sprintf(`
If you want to re-export a symbol, use the following pattern:

    load(..., _%s = %q, ...)
    %s = _%s
`, to.Name, from.Name, to.Name, to.Name)
				}
			}
		}

		if len(loadFindings) == 0 {
			// No problems with the current load statement
			continue
		}

		build.SortLoadArgs(&load)
		var newStmt build.Expr = &load
		if len(load.To) == 0 {
			// If there are no loaded symbols left remove the entire load statement
			newStmt = nil
		}
		replacement := LinterReplacement{&f.Stmt[stmtIndex], newStmt}

		// Individual replacements can't be combined together: assume we need to remove both loaded
		// symbols from
		//
		//     load(":foo.bzl", "a", "b")
		//
		// Individual replacements are just to remove each of the symbols, but if these replacements
		// are applied together, the result will be incorrect and a syntax error in Bazel:
		//
		//     load(":foo.bzl")
		//
		// A workaround is to attach the full replacement to the first finding.
		loadFindings[0].Replacement = []LinterReplacement{replacement}
		findings = append(findings, loadFindings...)
	}
	return findings
}

// collectLocalVariables traverses statements (e.g. of a function definition) and returns a list
// of idents for variables defined anywhere inside the function.
func collectLocalVariables(stmts []build.Expr) []*build.Ident {
	variables := []*build.Ident{}

	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *build.DefStmt:
			// Don't traverse nested functions
		case *build.ForStmt:
			variables = append(variables, bzlenv.CollectLValues(stmt.Vars)...)
			variables = append(variables, collectLocalVariables(stmt.Body)...)
		case *build.IfStmt:
			variables = append(variables, collectLocalVariables(stmt.True)...)
			variables = append(variables, collectLocalVariables(stmt.False)...)
		case *build.AssignExpr:
			variables = append(variables, bzlenv.CollectLValues(stmt.LHS)...)
		}
	}
	return variables
}

// searchUninitializedVariables takes a list of statements (e.g. body of a block statement)
// and a map of previously initialized statements, and calls `callback` on all idents that are not
// initialized. An ident is considered initialized if it's initialized by every possible execution
// path (before or by `stmts`).
// Returns a boolean indicating whether the current list of statements is guaranteed to be
// terminated explicitly (by return or fail() statements) and a map of variables that are guaranteed
// to be defined by `stmts`.
func findUninitializedVariables(stmts []build.Expr, previouslyInitialized map[string]bool, callback func(*build.Ident)) (bool, map[string]bool) {
	// Variables that are guaranteed to be initialized
	locallyInitialized := make(map[string]bool) // in the local block of `stmts`
	initialized := make(map[string]bool)        // anywhere before the current line
	for key := range previouslyInitialized {
		initialized[key] = true
	}

	// findUninitializedIdents traverses an expression (simple statement or a part of it), and calls
	// `callback` on every *build.Ident that's not mentioned in the map of initialized variables
	findUninitializedIdents := func(expr build.Expr, callback func(ident *build.Ident)) {
		// Collect lValues, they shouldn't be taken into account
		// For example, if the expression is `a = foo(b = c)`, only `c` can be an unused variable here.
		lValues := make(map[*build.Ident]bool)
		build.Walk(expr, func(expr build.Expr, stack []build.Expr) {
			if as, ok := expr.(*build.AssignExpr); ok {
				for _, ident := range bzlenv.CollectLValues(as.LHS) {
					lValues[ident] = true
				}
			}
		})

		build.Walk(expr, func(expr build.Expr, stack []build.Expr) {
			// TODO: traverse comprehensions properly
			for _, node := range stack {
				if _, ok := node.(*build.Comprehension); ok {
					return
				}
			}

			if ident, ok := expr.(*build.Ident); ok && !initialized[ident.Name] && !lValues[ident] {
				callback(ident)
			}
		})
	}

	for _, stmt := range stmts {
		newlyDefinedVariables := make(map[string]bool)
		switch stmt := stmt.(type) {
		case *build.DefStmt:
			// Don't traverse nested functions
		case *build.CallExpr:
			if _, ok := isFunctionCall(stmt, "fail"); ok {
				return true, locallyInitialized
			}
		case *build.ReturnStmt:
			findUninitializedIdents(stmt, callback)
			return true, locallyInitialized
		case *build.BranchStmt:
			if stmt.Token == "break" || stmt.Token == "continue" {
				return true, locallyInitialized
			}
		case *build.ForStmt:
			// Although loop variables are defined as local variables, buildifier doesn't know whether
			// the collection will be empty or not.

			// Traverse but ignore the result. Even if something is defined inside a for-loop, the loop
			// may be empty and the variable initialization may not happen.
			findUninitializedIdents(stmt.X, callback)

			// The loop can access the variables defined above, and also the for-loop variables.
			initializedInLoop := make(map[string]bool)
			for name := range initialized {
				initializedInLoop[name] = true
			}
			for _, ident := range bzlenv.CollectLValues(stmt.Vars) {
				initializedInLoop[ident.Name] = true
			}
			findUninitializedVariables(stmt.Body, initializedInLoop, callback)
			continue
		case *build.IfStmt:
			findUninitializedIdents(stmt.Cond, callback)
			// Check the variables defined in the if- and else-clauses.
			terminatedTrue, definedInTrue := findUninitializedVariables(stmt.True, initialized, callback)
			terminatedFalse, definedInFalse := findUninitializedVariables(stmt.False, initialized, callback)
			if terminatedTrue && terminatedFalse {
				return true, locallyInitialized
			} else if terminatedTrue {
				// Only take definedInFalse into account
				for key := range definedInFalse {
					locallyInitialized[key] = true
					initialized[key] = true
				}
			} else if terminatedFalse {
				// Only take definedInTrue into account
				for key := range definedInTrue {
					locallyInitialized[key] = true
					initialized[key] = true
				}
			} else {
				// If a variable is defined in both if- and else-clauses, it's considered as defined
				for key := range definedInTrue {
					if definedInFalse[key] {
						locallyInitialized[key] = true
						initialized[key] = true
					}
				}
			}
			continue
		case *build.AssignExpr:
			// Assignment expression. Collect all definitions from the lhs
			for _, ident := range bzlenv.CollectLValues(stmt.LHS) {
				newlyDefinedVariables[ident.Name] = true
			}
		}
		findUninitializedIdents(stmt, callback)
		for name := range newlyDefinedVariables {
			locallyInitialized[name] = true
			initialized[name] = true
		}
	}
	return false, locallyInitialized
}

// uninitializedVariableWarning warns about usages of values that may not have been initialized.
func uninitializedVariableWarning(f *build.File) []*LinterFinding {
	findings := []*LinterFinding{}
	for _, stmt := range f.Stmt {
		def, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		// Get all variables defined in the function body.
		// If a variable is not defined there, it can be builtin, global, or loaded.
		localVars := make(map[string]bool)
		for _, ident := range collectLocalVariables(def.Body) {
			localVars[ident.Name] = true
		}

		// Function parameters are guaranteed to be defined everywhere in the function, even if they
		// are redefined inside the function body. They shouldn't be taken into consideration.
		for _, param := range def.Params {
			if name, _ := build.GetParamName(param); name != "" {
				delete(localVars, name)
			}
		}

		// Search for all potentially initialized variables in the function body
		findUninitializedVariables(def.Body, make(map[string]bool), func(ident *build.Ident) {
			// Check that the found ident represents a local variable
			if localVars[ident.Name] {
				findings = append(findings,
					makeLinterFinding(ident, fmt.Sprintf(`Variable "%s" may not have been initialized.`, ident.Name)))
			}
		})
	}
	return findings
}
