// Cosmetic warnings (e.g. for improved readability of Starlark files)

package warn

import (
	"sort"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
	"github.com/bazelbuild/buildtools/tables"
)

func sameOriginLoadWarning(f *build.File, fix bool) []*Finding {
	findings := []*Finding{}
	loaded := make(map[string]*build.LoadStmt)
	for stmtIndex := 0; stmtIndex < len(f.Stmt); stmtIndex++ {
		load, ok := f.Stmt[stmtIndex].(*build.LoadStmt)
		if !ok {
			continue
		}

		if fix {
			start, end := load.Span()
			if start.Line == end.Line {
				load.ForceCompact = true
			}
		}

		previousLoad := loaded[load.Module.Value]
		if previousLoad == nil {
			loaded[load.Module.Value] = load
			continue
		}

		if fix {
			previousLoad.To = append(previousLoad.To, load.To...)
			previousLoad.From = append(previousLoad.From, load.From...)

			// Force the merged load statement to be compact if both previous and current load statements are compact
			if !load.ForceCompact {
				previousLoad.ForceCompact = false
			}

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

	if f.Type != build.TypeDefault {
		// Only applicable to .bzl files
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
			// Either both labels have explicit repository names or both don't, compare their packages
			// and break ties using file names if necessary

			module1Parts := strings.Split(strings.TrimLeft(load1Label, "@"), ":")
			package1, filename1 := module1Parts[0], module1Parts[1]
			module2Parts := strings.Split(strings.TrimLeft(load2Label, "@"), ":")
			package2, filename2 := module2Parts[0], module2Parts[1]

			// in case both packages are the same, use file names to break ties
			if package1 == package2 {
				return filename1 < filename2
			}

			// in case one of the packages is empty, the empty one goes first
			if len(package1) == 0 || len(package2) == 0 {
				return len(package1) > 0
			}

			// both packages are non-empty and not equal, so compare them
			return package1 < package2
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
