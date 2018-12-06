// Package warn implements functions that generate warnings for BUILD files.
package warn

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
	"github.com/bazelbuild/buildtools/tables"
)

// A Finding is a warning reported by the analyzer. It may contain an optional suggested fix.
type Finding struct {
	File        *build.File
	Start       build.Position
	End         build.Position
	Category    string
	Message     string
	URL         string
	Actionable  bool
	Replacement *Replacement
}

// A Replacement is a suggested fix. Text between Start and End should be replaced with Content.
type Replacement struct {
	Description string
	Start       build.Position
	End         build.Position
	Content     string
}

var functionsWithPositionalArguments = map[string]bool{
	"distribs":            true,
	"exports_files":       true,
	"licenses":            true,
	"print":               true,
	"register_toolchains": true,
	"vardef":              true,
}

func docURL(cat string) string {
	return "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#" + cat
}

// makeFinding creates a Finding object
func makeFinding(f *build.File, start, end build.Position, cat string, msg string, actionable bool, fix *Replacement) *Finding {
	return &Finding{
		File:        f,
		Start:       start,
		End:         end,
		Category:    cat,
		URL:         docURL(cat),
		Message:     msg,
		Actionable:  actionable,
		Replacement: fix,
	}
}

// MakeFix creates a Replacement object
func MakeFix(f *build.File, desc string, start build.Position, end build.Position, newContent string) *Replacement {
	return &Replacement{
		Description: desc,
		Start:       start,
		End:         end,
		Content:     newContent,
	}
}

func positionalArgumentsWarning(f *build.File, pkg string, stmt build.Expr) *Finding {
	msg := "All calls to rules or macros should pass arguments by keyword (arg_name=value) syntax."
	call, ok := stmt.(*build.CallExpr)
	if !ok {
		return nil
	}
	if id, ok := call.X.(*build.Ident); !ok || functionsWithPositionalArguments[id.Name] {
		return nil
	}
	for _, arg := range call.List {
		if op, ok := arg.(*build.BinaryExpr); ok && op.Op == "=" {
			continue
		}
		start, end := arg.Span()
		return makeFinding(f, start, end, "positional-args", msg, true, nil)
	}
	return nil
}

func constantGlobWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	edit.EditFunction(f, "glob", func(call *build.CallExpr, stk []build.Expr) build.Expr {
		if len(call.List) == 0 {
			return nil
		}
		patterns, ok := call.List[0].(*build.ListExpr)
		if !ok {
			return nil
		}
		for _, expr := range patterns.List {
			str, ok := expr.(*build.StringExpr)
			if !ok {
				continue
			}
			if !strings.Contains(str.Value, "*") {
				start, end := str.Span()
				findings = append(findings, makeFinding(f, start, end, "constant-glob",
					"Glob pattern `"+str.Value+"` has no wildcard ('*'). "+
						"Constant patterns can be error-prone, move the file outside the glob.", true, nil))
				return nil // at most one warning per glob
			}
		}
		return nil
	})
	return findings
}

func unusedLoadWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	loaded := make(map[string]bool)

	symbols := edit.UsedSymbols(f)
	// statements are considered in reserve order to make sure that the last
	// load wins, just like what happens during evaluation, which is very
	// important if same symbol is loaded from different files.
	for stmtIndex := len(f.Stmt) - 1; stmtIndex >= 0; stmtIndex-- {
		load, ok := f.Stmt[stmtIndex].(*build.LoadStmt)
		if !ok {
			continue
		}
		for i := 0; i < len(load.To); i++ {
			from := load.From[i]
			to := load.To[i]
			// Check if the symbol was already loaded
			if loaded[to.Name] {
				if fix {
					load.To = append(load.To[:i], load.To[i+1:]...)
					load.From = append(load.From[:i], load.From[i+1:]...)
					i--
				} else {
					start, end := to.Span()
					findings = append(findings,
						makeFinding(f, start, end, "load",
							"Symbol \""+to.Name+"\" has already been loaded. Please remove it.", true, nil))
				}
				continue
			}
			_, ok := symbols[to.Name]
			if !ok && !edit.ContainsComments(load, "@unused") && !edit.ContainsComments(to, "@unused") && !edit.ContainsComments(from, "@unused") {
				// To disable the warning, put a comment that contains '@unused'
				if fix {
					load.To = append(load.To[:i], load.To[i+1:]...)
					load.From = append(load.From[:i], load.From[i+1:]...)
					i--
				} else {
					start, end := to.Span()
					findings = append(findings,
						makeFinding(f, start, end, "load",
							"Loaded symbol \""+to.Name+"\" is unused. Please remove it.\n"+
								"To disable the warning, add '@unused' in a comment.", true, nil))

				}
			}
			loaded[to.Name] = true
		}
		// If there are no loaded symbols left remove the entire load statement
		if fix && len(load.To) == 0 {
			f.Stmt = append(f.Stmt[:stmtIndex], f.Stmt[stmtIndex+1:]...)
		}
	}
	return findings
}

func sameOriginLoadWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	loaded := make(map[string]*build.LoadStmt)
	for stmtIndex := 0; stmtIndex < len(f.Stmt); stmtIndex++ {
		load, ok := f.Stmt[stmtIndex].(*build.LoadStmt)
		if !ok {
			continue
		}

		previousLoad := loaded[load.Module.Value]
		if previousLoad == nil {
			loaded[load.Module.Value] = load
			continue
		}

		if fix {
			previousLoad.To = append(previousLoad.To, load.To...)
			previousLoad.From = append(previousLoad.From, load.From...)
			f.Stmt = append(f.Stmt[:stmtIndex], f.Stmt[stmtIndex+1:]...)
			stmtIndex--
		} else {
			start, end := load.Module.Span()
			findings = append(findings,
				makeFinding(f, start, end, "same-origin-load",
					"There is already a load from \""+load.Module.Value+"\". Please merge all loads from the same origin into a single one.", true, nil))
		}
	}
	return findings
}

func redefinedVariableWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	definedSymbols := make(map[string]bool)

	for _, s := range f.Stmt {
		// look for all assignments in the scope
		as, ok := s.(*build.BinaryExpr)
		if !ok || as.Op != "=" {
			continue
		}
		start, end := as.X.Span()
		left, ok := as.X.(*build.Ident)
		if !ok {
			continue
		}
		if definedSymbols[left.Name] {
			findings = append(findings,
				makeFinding(f, start, end, "redefined-variable",
					"Variable \""+left.Name+"\" has already been defined. "+
						"Redefining a global value is discouraged and will be forbidden in the future.\n"+
						"Consider using a new variable instead.", true, nil))
			continue
		}
		definedSymbols[left.Name] = true
	}
	return findings
}

func unusedVariableWarning(f *build.File, fix bool) []*Finding {
	return unusedVariableCheck(f, f.Stmt, []*Finding{})
}

func duplicatedNameWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	if f.Type == build.TypeDefault {
		// Not applicable to .bzl files.
		return findings
	}
	names := make(map[string]int) // map from name to line number
	msg := "A rule with name `%s' was already found on line %d. " +
		"Even if it's valid for Blaze, this may confuse other tools. " +
		"Please rename it and use different names."

	for _, rule := range f.Rules("") {
		name := rule.Name()
		if name == "" {
			continue
		}
		start, end := rule.Call.Span()
		if nameNode := rule.Attr("name"); nameNode != nil {
			start, end = nameNode.Span()
		}
		if line, ok := names[name]; ok {
			findings = append(findings,
				makeFinding(f, start, end, "duplicated-name", fmt.Sprintf(msg, name, line), true, nil))
		} else {
			names[name] = start.Line
		}
	}
	return findings
}

func packageOnTopWarning(f *build.File, fix bool) []*Finding {
	seenRule := false
	for _, stmt := range f.Stmt {
		_, isString := stmt.(*build.StringExpr) // typically a docstring
		_, isComment := stmt.(*build.CommentBlock)
		_, isBinaryExpr := stmt.(*build.BinaryExpr) // e.g. variable declaration
		_, isLoad := stmt.(*build.LoadStmt)
		_, isPackageGroup := edit.ExprToRule(stmt, "package_group")
		_, isLicense := edit.ExprToRule(stmt, "licenses")
		if isString || isComment || isBinaryExpr || isLoad || isPackageGroup || isLicense {
			continue
		}
		if rule, ok := edit.ExprToRule(stmt, "package"); ok {
			if !seenRule { // OK: package is on top of the file
				return nil
			}
			start, end := rule.Call.Span()
			return []*Finding{makeFinding(f, start, end, "package-on-top",
				"Package declaration should be at the top of the file, after the load() statements, "+
					"but before any call to a rule or a macro. "+
					"package_group() and licenses() may be called before package().", true, nil)}
		}
		seenRule = true
	}
	return nil
}

func loadOnTopWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	if f.Type == build.TypeWorkspace {
		// Not applicable for WORKSPACE files
		return findings
	}

	firstStmtIndex := -1 // index of the first seen non-load statement
	for i := 0; i < len(f.Stmt); i++ {
		stmt := f.Stmt[i]
		_, isString := stmt.(*build.StringExpr) // typically a docstring
		_, isComment := stmt.(*build.CommentBlock)
		if isString || isComment {
			continue
		}
		load, ok := stmt.(*build.LoadStmt)
		if !ok {
			if firstStmtIndex == -1 {
				firstStmtIndex = i
			}
			continue
		}
		if firstStmtIndex == -1 {
			continue
		}
		if !fix {
			start, end := load.Span()
			findings = append(findings, makeFinding(f, start, end, "load-on-top",
				"Load statements should be at the top of the file.", true, nil))
			continue
		}
		stmts := []build.Expr{}
		stmts = append(stmts, f.Stmt[:firstStmtIndex]...)
		stmts = append(stmts, load)
		stmts = append(stmts, f.Stmt[firstStmtIndex:i]...)
		stmts = append(stmts, f.Stmt[i+1:]...)
		f.Stmt = stmts
		firstStmtIndex++
	}
	return findings
}

func outOfOrderLoadWarning(f *build.File, fix bool) []*Finding {
	// compareLoadLabels compares two module names
	compareLoadLabels := func(load1Label, load2Label string) bool {
		// handle absolute labels with explicit repositories separately to
		// make sure they preceed absolute and relative labels without repos
		isExplicitRepo1 := strings.HasPrefix(load1Label, "@")
		isExplicitRepo2 := strings.HasPrefix(load2Label, "@")
		if isExplicitRepo1 == isExplicitRepo2 {
			// Either both labels have explicit repository names or both don't, compare lexicographically
			return load1Label < load2Label
		}
		// Exactly one label has an explicit repository name, it should be the first one.
		return isExplicitRepo1
	}

	findings := []*Finding{}

	if f.Type == build.TypeWorkspace {
		// Not applicable for WORKSPACE files
		return findings
	}

	sortedLoads := []*build.LoadStmt{}
	for i := 0; i < len(f.Stmt); i++ {
		load, ok := f.Stmt[i].(*build.LoadStmt)
		if !ok {
			continue
		}
		sortedLoads = append(sortedLoads, load)
	}
	if fix {
		sort.SliceStable(sortedLoads, func(i, j int) bool {
			load1Label := sortedLoads[i].Module.Value
			load2Label := sortedLoads[j].Module.Value
			return compareLoadLabels(load1Label, load2Label)
		})
		sortedLoadIndex := 0
		for globalLoadIndex := 0; globalLoadIndex < len(f.Stmt); globalLoadIndex++ {
			if _, ok := f.Stmt[globalLoadIndex].(*build.LoadStmt); !ok {
				continue
			}
			f.Stmt[globalLoadIndex] = sortedLoads[sortedLoadIndex]
			sortedLoadIndex++
		}
		return findings
	}

	for i := 1; i < len(sortedLoads); i++ {
		if compareLoadLabels(sortedLoads[i].Module.Value, sortedLoads[i-1].Module.Value) {
			start, end := sortedLoads[i].Span()
			findings = append(findings, makeFinding(f, start, end, "out-of-order-load",
				"Load statement is out of its lexicographical order.", true, nil))
		}
	}
	return findings
}

func integerDivisionWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		binary, ok := expr.(*build.BinaryExpr)
		if !ok {
			return
		}
		if binary.Op != "/" && binary.Op != "/=" {
			return
		}
		if fix {
			binary.Op = "/" + binary.Op
			return
		}
		start, end := binary.Span()
		findings = append(findings,
			makeFinding(f, start, end, "integer-division",
				"The \""+binary.Op+"\" operator for integer division is deprecated in favor of \"/"+binary.Op+"\".", true, nil))
	})
	return findings
}

func unsortedDictItemsWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	compareItems := func(item1, item2 *build.KeyValueExpr) bool {
		key1 := item1.Key.(*build.StringExpr).Value
		key2 := item2.Key.(*build.StringExpr).Value
		// regular keys should preceed private ones (start with "_")
		if strings.HasPrefix(key1, "_") {
			return strings.HasPrefix(key2, "_") && key1 < key2
		}
		if strings.HasPrefix(key2, "_") {
			return true
		}
		key1Priority := tables.NamePriority[key1]
		key2Priority := tables.NamePriority[key2]
		if key1Priority != key2Priority {
			return key1Priority < key2Priority
		}
		return key1 < key2
	}

	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		dict, ok := expr.(*build.DictExpr)

		mustSkipCheck := func(expr build.Expr) bool {
			return edit.ContainsComments(expr, "@unsorted-dict-items")
		}

		if !ok || mustSkipCheck(dict) {
			return
		}
		// do not process dictionaries nested within expressions that do not
		// want dict items to be sorted
		for i := len(stack) - 1; i >= 0; i-- {
			if mustSkipCheck(stack[i]) {
				return
			}
		}
		sortedItems := []*build.KeyValueExpr{}
		for _, stmt := range dict.List {
			item, ok := stmt.(*build.KeyValueExpr)
			if !ok {
				continue
			}
			// include only string literal keys into consideration
			if _, ok = item.Key.(*build.StringExpr); !ok {
				continue
			}
			sortedItems = append(sortedItems, item)
		}
		if fix {
			sort.SliceStable(sortedItems, func(i, j int) bool {
				return compareItems(sortedItems[i], sortedItems[j])
			})
			sortedItemIndex := 0
			for originalItemIndex := 0; originalItemIndex < len(dict.List); originalItemIndex++ {
				item, ok := dict.List[originalItemIndex].(*build.KeyValueExpr)
				if !ok {
					continue
				}
				if _, ok := item.Key.(*build.StringExpr); !ok {
					continue
				}
				dict.List[originalItemIndex] = sortedItems[sortedItemIndex]
				sortedItemIndex++
			}
			return
		}

		for i := 1; i < len(sortedItems); i++ {
			if compareItems(sortedItems[i], sortedItems[i-1]) {
				start, end := sortedItems[i].Span()
				findings = append(findings, makeFinding(f, start, end, "unsorted-dict-items",
					"Dictionary items are out of their lexicographical order.", true, nil))
			}
		}
		return
	})
	return findings
}

func isBranchStmt(e build.Expr) bool {
	// TODO(laurentlb): This should be a separate node in the AST.
	if id, ok := e.(*build.Ident); ok {
		if id.Name == "break" || id.Name == "continue" || id.Name == "pass" {
			return true
		}
	}
	return false
}

func noEffectStatementsCheck(f *build.File, body []build.Expr, isTopLevel, isFunc bool, findings []*Finding) []*Finding {
	seenNonComment := false
	for _, stmt := range body {
		start, end := stmt.Span()
		if _, ok := stmt.(*build.StringExpr); ok {
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
			*build.CallExpr, *build.CommentBlock:
			continue
		case *build.BinaryExpr:
			if s.Op != "==" && s.Op != "!=" && strings.HasSuffix(s.Op, "=") {
				continue
			}
		}
		if isBranchStmt(stmt) {
			continue
		}
		if comp, ok := stmt.(*build.Comprehension); ok {
			if !isTopLevel || comp.Curly {
				// List comprehensions are allowed on top-level.
				findings = append(findings,
					makeFinding(f, start, end, "no-effect",
						"Expression result is not used. Use a for-loop instead of a list comprehension.", true, nil))
			}
			continue
		}
		findings = append(findings,
			makeFinding(f, start, end, "no-effect",
				"Expression result is not used.", true, nil))
	}
	return findings
}

// unusedVariableCheck checks for unused variables inside a given node `stmt` (either *build.File or
// *build.DefStmt) and reports unused and already defined variables.
func unusedVariableCheck(f *build.File, stmts []build.Expr, findings []*Finding) []*Finding {
	if f.Type == build.TypeDefault {
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
		as, ok := s.(*build.BinaryExpr)
		if !ok || as.Op != "=" {
			continue
		}
		start, end := as.X.Span()
		left, ok := as.X.(*build.Ident)
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
			makeFinding(f, start, end, "unused-variable",
				"Variable \""+left.Name+"\" is unused. Please remove it.\n"+
					"To disable the warning, add '@unused' in a comment.", true, nil))
	}
	return findings
}

func noEffectWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	findings = noEffectStatementsCheck(f, f.Stmt, true, false, findings)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		// The AST should have a ExprStmt node.
		// Since we don't have that, we match on the nodes that contain a block to get the list of statements.
		switch expr := expr.(type) {
		case *build.ForStmt:
			findings = noEffectStatementsCheck(f, expr.Body, false, false, findings)
		case *build.DefStmt:
			findings = noEffectStatementsCheck(f, expr.Function.Body, false, true, findings)
		case *build.IfStmt:
			findings = noEffectStatementsCheck(f, expr.True, false, false, findings)
			findings = noEffectStatementsCheck(f, expr.False, false, false, findings)
		}
	})
	return findings
}

func dictionaryConcatenationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		binary, ok := expr.(*build.BinaryExpr)
		if !ok {
			return
		}
		if binary.Op != "+" && binary.Op != "+=" {
			return
		}
		if types[binary.X] == Dict || types[binary.Y] == Dict {
			start, end := binary.Span()
			findings = append(findings,
				makeFinding(f, start, end, "dict-concatenation",
					"Dictionary concatenation is deprecated.", true, nil))
		}
	})
	return findings
}

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

func stringIterationWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	addWarning := func(expr build.Expr) {
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "string-iteration",
				"String iteration is deprecated.", true, nil))
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

func nativeInBuildFilesWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	if f.Type != build.TypeBuild {
		return findings
	}

	build.Edit(f, func(expr build.Expr, stack []build.Expr) build.Expr {
		// Search for `native.xxx` nodes
		dot, ok := expr.(*build.DotExpr)
		if !ok {
			return nil
		}
		ident, ok := dot.X.(*build.Ident)
		if !ok || ident.Name != "native" {
			return nil
		}

		if fix {
			start, _ := dot.Span()
			return &build.Ident{
				Name:    dot.Name,
				NamePos: start,
			}
		}
		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "native-build",
				`The "native" module shouldn't be used in BUILD files, its members are available as global symbols.`, true, nil))

		return nil
	})
	return findings
}

func nativePackageWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}

	if f.Type != build.TypeDefault {
		return findings
	}

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

		start, end := expr.Span()
		findings = append(findings,
			makeFinding(f, start, end, "native-package",
				`"native.package()" shouldn't be used in .bzl files.`, true, nil))
	})
	return findings
}

// RuleWarningMap lists the warnings that run on a single rule.
// These warnings run only on BUILD files (not bzl files).
var RuleWarningMap = map[string]func(f *build.File, pkg string, expr build.Expr) *Finding{
	"positional-args": positionalArgumentsWarning,
}

// FileWarningMap lists the warnings that run on the whole file.
var FileWarningMap = map[string]func(f *build.File, fix bool) []*Finding{
	"args-order":          argumentsOrderWarning,
	"attr-cfg":            attrConfigurationWarning,
	"attr-license":        attrLicenseWarning,
	"attr-non-empty":      attrNonEmptyWarning,
	"attr-output-default": attrOutputDefaultWarning,
	"attr-single-file":    attrSingleFileWarning,
	"constant-glob":       constantGlobWarning,
	"ctx-actions":         ctxActionsWarning,
	"ctx-args":            contextArgsAPIWarning,
	"depset-iteration":    depsetIterationWarning,
	"depset-union":        depsetUnionWarning,
	"dict-concatenation":  dictionaryConcatenationWarning,
	"duplicated-name":     duplicatedNameWarning,
	"filetype":            fileTypeWarning,
	"git-repository":      nativeGitRepositoryWarning,
	"http-archive":        nativeHTTPArchiveWarning,
	"integer-division":    integerDivisionWarning,
	"load":                unusedLoadWarning,
	"load-on-top":         loadOnTopWarning,
	"native-build":        nativeInBuildFilesWarning,
	"native-package":      nativePackageWarning,
	"no-effect":           noEffectWarning,
	"out-of-order-load":   outOfOrderLoadWarning,
	"output-group":        outputGroupWarning,
	"package-name":        packageNameWarning,
	"package-on-top":      packageOnTopWarning,
	"redefined-variable":  redefinedVariableWarning,
	"repository-name":     repositoryNameWarning,
	"same-origin-load":    sameOriginLoadWarning,
	"string-iteration":    stringIterationWarning,
	"unsorted-dict-items": unsortedDictItemsWarning,
	"unused-variable":     unusedVariableWarning,
}

// DisabledWarning checks if the warning was disabled by a comment.
// The comment format is buildozer: disable=<warning>
func DisabledWarning(f *build.File, finding *Finding, warning string) bool {
	format := "buildozer: disable=" + warning
	findingLine := finding.Start.Line

	for _, stmt := range f.Stmt {
		stmtStart, _ := stmt.Span()
		if stmtStart.Line == findingLine {
			// Is this specific line disabled?
			if edit.ContainsComments(stmt, format) {
				return true
			}
		}
		// Check comments within a rule
		rule, ok := stmt.(*build.CallExpr)
		if ok {
			for _, stmt := range rule.List {
				stmtStart, _ := stmt.Span()
				if stmtStart.Line != findingLine {
					continue
				}
				// Is the whole rule or this specific line as a comment
				// to disable this warning?
				if edit.ContainsComments(rule, format) ||
					edit.ContainsComments(stmt, format) {
					return true
				}
			}
		}
		// Check comments within a load statement
		load, ok := stmt.(*build.LoadStmt)
		if ok {
			loadHasComment := edit.ContainsComments(load, format)
			module := load.Module
			if module.Start.Line == findingLine {
				if edit.ContainsComments(module, format) || loadHasComment {
					return true
				}
			}
			for i, to := range load.To {
				from := load.From[i]
				if to.NamePos.Line == findingLine || from.NamePos.Line == findingLine {
					if edit.ContainsComments(to, format) || edit.ContainsComments(from, format) || loadHasComment {
						return true
					}
				}
			}
		}
	}

	return false
}

// FileWarnings returns a list of all warnings found in the file.
func FileWarnings(f *build.File, pkg string, enabledWarnings []string, fix bool) []*Finding {
	findings := []*Finding{}
	for _, warn := range enabledWarnings {
		if fct, ok := FileWarningMap[warn]; ok {
			for _, w := range fct(f, fix) {
				if !DisabledWarning(f, w, warn) {
					findings = append(findings, w)
				}
			}
		} else {
			fn := RuleWarningMap[warn]
			if fn == nil {
				log.Fatalf("unexpected warning %q", warn)
			}
			if f.Type == build.TypeDefault {
				continue
			}
			for _, stmt := range f.Stmt {
				if w := fn(f, pkg, stmt); w != nil {
					if !DisabledWarning(f, w, warn) {
						findings = append(findings, w)
					}
				}
			}
		}
	}
	return findings
}

// PrintWarnings prints the list of warnings returned from calling FileWarnings.
// Actionable warnings list their link in parens, inactionable warnings list
// their link in square brackets.
func PrintWarnings(f *build.File, pkg string, enabledWarnings []string, showReplacements bool) bool {
	warnings := FileWarnings(f, pkg, enabledWarnings, false)
	sort.Slice(warnings, func(i, j int) bool { return warnings[i].Start.Line < warnings[j].Start.Line })
	for _, w := range warnings {
		formatString := "%s:%d: %s: %s (%s)"
		if !w.Actionable {
			formatString = "%s:%d: %s: %s [%s]"
		}
		fmt.Fprintf(os.Stderr, formatString,
			w.File.Path,
			w.Start.Line,
			w.Category,
			w.Message,
			w.URL)
		if showReplacements && w.Replacement != nil {
			r := w.Replacement
			fmt.Fprintf(os.Stderr, " [%d..%d): %s\n",
				r.Start.Byte,
				r.End.Byte,
				r.Content)
		} else {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	return len(warnings) > 0
}

// FixWarnings fixes all warnings that can be fixed automatically.
func FixWarnings(f *build.File, pkg string, enabledWarnings []string, verbose bool) {
	warnings := FileWarnings(f, pkg, enabledWarnings, true)
	if verbose {
		fmt.Fprintf(os.Stderr, "%s: applied fixes, %d warnings left\n",
			f.Path,
			len(warnings))
	}
}

func collectAllWarnings() []string {
	var result []string
	// Collect list of all warnings.
	for k := range FileWarningMap {
		result = append(result, k)
	}
	for k := range RuleWarningMap {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// AllWarnings is the list of all available warnings.
var AllWarnings = collectAllWarnings()
