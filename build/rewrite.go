/*
Copyright 2016 Google LLC

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

// Rewriting of high-level (not purely syntactic) BUILD constructs.

package build

import (
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bazelbuild/buildtools/tables"
)

// DisableRewrites disables certain rewrites (for debugging).
var DisableRewrites []string

// disabled reports whether the named rewrite is disabled.
func disabled(name string) bool {
	for _, x := range DisableRewrites {
		if name == x {
			return true
		}
	}
	return false
}

// AllowSort allows sorting of these lists even with sorting otherwise disabled (for debugging).
var AllowSort []string

// allowedSort reports whether sorting is allowed in the named context.
func allowedSort(name string) bool {
	for _, x := range AllowSort {
		if name == x {
			return true
		}
	}
	return false
}

// Rewriter controls the rewrites to be applied.
//
// If non-nil, the rewrites with the specified names will be run. If
// nil, a default set of rewrites will be used that is determined by
// the type (BUILD vs default starlark) of the file being rewritten.
type Rewriter struct {
	RewriteSet                      []string
	IsLabelArg                      map[string]bool
	LabelDenyList                   map[string]bool
	IsSortableListArg               map[string]bool
	SortableDenylist                map[string]bool
	SortableAllowlist               map[string]bool
	NamePriority                    map[string]int
	StripLabelLeadingSlashes        bool
	ShortenAbsoluteLabelsToRelative bool
}

// Rewrite applies rewrites to a file
func Rewrite(f *File) {
	var rewriter = &Rewriter{
		IsLabelArg:                      tables.IsLabelArg,
		LabelDenyList:                   tables.LabelDenylist,
		IsSortableListArg:               tables.IsSortableListArg,
		SortableDenylist:                tables.SortableDenylist,
		SortableAllowlist:               tables.SortableAllowlist,
		NamePriority:                    tables.NamePriority,
		StripLabelLeadingSlashes:        tables.StripLabelLeadingSlashes,
		ShortenAbsoluteLabelsToRelative: tables.ShortenAbsoluteLabelsToRelative,
	}
	rewriter.Rewrite(f)
}

// Rewrite applies the rewrites to a file
func (w *Rewriter) Rewrite(f *File) {
	for _, r := range rewrites {
		// f.Type&r.scope is a bitwise comparison. Because starlark files result in a scope that will
		// not be changed by rewrites, we have included another check looking on the right side.
		// If we have an empty rewrite set, we do not want any rewrites to happen.
		if (!disabled(r.name) && (f.Type&r.scope != 0) && w.RewriteSet == nil) || (w.RewriteSet != nil && rewriteSetContains(w, r.name)) {
			r.fn(f, w)
		}
	}
}

func rewriteSetContains(w *Rewriter, name string) bool {
	for _, value := range w.RewriteSet {
		if value == name {
			return true
		}
	}
	return false
}

// Each rewrite function can be either applied for BUILD files, other files (such as .bzl),
// or all files.
const (
	scopeDefault = TypeDefault | TypeBzl                  // .bzl and generic Starlark files
	scopeBuild   = TypeBuild | TypeWorkspace | TypeModule // BUILD, WORKSPACE, and MODULE files
	scopeBoth    = scopeDefault | scopeBuild
)

// rewrites is the list of all Buildifier rewrites, in the order in which they are applied.
// The order here matters: for example, label canonicalization must happen
// before sorting lists of strings.
var rewrites = []struct {
	name  string
	fn    func(*File, *Rewriter)
	scope FileType
}{
	{"removeParens", removeParens, scopeBuild},
	{"callsort", sortCallArgs, scopeBuild},
	{"label", fixLabels, scopeBuild},
	{"listsort", sortStringLists, scopeBoth},
	{"multiplus", fixMultilinePlus, scopeBuild},
	{"loadTop", moveLoadOnTop, scopeBoth},
	{"sameOriginLoad", compressSameOriginLoads, scopeBoth},
	{"sortLoadStatements", sortLoadStatements, scopeBoth},
	{"loadsort", sortAllLoadArgs, scopeBoth},
	{"useRepoPositionalsSort", sortUseRepoPositionals, TypeModule},
	{"formatdocstrings", formatDocstrings, scopeBoth},
	{"reorderarguments", reorderArguments, scopeBoth},
	{"editoctal", editOctals, scopeBoth},
}

// leaveAlone reports whether any of the nodes on the stack are marked
// with a comment containing "buildifier: leave-alone".
func leaveAlone(stk []Expr, final Expr) bool {
	for _, x := range stk {
		if leaveAlone1(x) {
			return true
		}
	}
	if final != nil && leaveAlone1(final) {
		return true
	}
	return false
}

// hasComment reports whether x is marked with a comment that
// after being converted to lower case, contains the specified text.
func hasComment(x Expr, text string) bool {
	if x == nil {
		return false
	}
	for _, com := range x.Comment().Before {
		if strings.Contains(strings.ToLower(com.Token), text) {
			return true
		}
	}
	for _, com := range x.Comment().After {
		if strings.Contains(strings.ToLower(com.Token), text) {
			return true
		}
	}
	for _, com := range x.Comment().Suffix {
		if strings.Contains(strings.ToLower(com.Token), text) {
			return true
		}
	}
	return false
}

// isCommentAnywhere checks whether there's a comment containing the given text
// anywhere in the file.
func isCommentAnywhere(f *File, text string) bool {
	commentExists := false
	WalkInterruptable(f, func(node Expr, stack []Expr) (err error) {
		if hasComment(node, text) {
			commentExists = true
			return &StopTraversalError{}
		}
		return nil
	})
	return commentExists
}

// leaveAlone1 reports whether x is marked with a comment containing
// "buildifier: leave-alone", case-insensitive.
func leaveAlone1(x Expr) bool {
	return hasComment(x, "buildifier: leave-alone")
}

// doNotSort reports whether x is marked with a comment containing
// "do not sort", case-insensitive.
func doNotSort(x Expr) bool {
	return hasComment(x, "do not sort")
}

// keepSorted reports whether x is marked with a comment containing
// "keep sorted", case-insensitive.
func keepSorted(x Expr) bool {
	return hasComment(x, "keep sorted")
}

// fixLabels rewrites labels into a canonical form.
//
// First, it joins labels written as string addition, turning
// "//x" + ":y" (usually split across multiple lines) into "//x:y".
//
// Second, it removes redundant target qualifiers, turning labels like
// "//third_party/m4:m4" into "//third_party/m4" as well as ones like
// "@foo//:foo" into "@foo".
func fixLabels(f *File, w *Rewriter) {
	joinLabel := func(p *Expr) {
		add, ok := (*p).(*BinaryExpr)
		if !ok || add.Op != "+" {
			return
		}
		str1, ok := add.X.(*StringExpr)
		if !ok || !strings.HasPrefix(str1.Value, "//") || strings.Contains(str1.Value, " ") {
			return
		}
		str2, ok := add.Y.(*StringExpr)
		if !ok || strings.Contains(str2.Value, " ") {
			return
		}
		str1.Value += str2.Value

		// Deleting nodes add and str2.
		// Merge comments from add, str1, and str2 and save in str1.
		com1 := add.Comment()
		com2 := str1.Comment()
		com3 := str2.Comment()
		com1.Before = append(com1.Before, com2.Before...)
		com1.Before = append(com1.Before, com3.Before...)
		com1.Suffix = append(com1.Suffix, com2.Suffix...)
		com1.Suffix = append(com1.Suffix, com3.Suffix...)
		*str1.Comment() = *com1

		*p = str1
	}

	labelPrefix := "//"
	if w.StripLabelLeadingSlashes {
		labelPrefix = ""
	}
	// labelRE matches label strings, e.g. @r//x/y/z:abc
	// where $1 is @r//x/y/z, $2 is @r//, $3 is r, $4 is z, $5 is abc.
	labelRE := regexp.MustCompile(`^(((?:@(\w+))?//|` + labelPrefix + `)(?:.+/)?([^:]*))(?::([^:]+))?$`)

	shortenLabel := func(v Expr) {
		str, ok := v.(*StringExpr)
		if !ok {
			return
		}

		if w.StripLabelLeadingSlashes && strings.HasPrefix(str.Value, "//") {
			if filepath.Dir(f.Path) == "." || !strings.HasPrefix(str.Value, "//:") {
				str.Value = str.Value[2:]
			}
		}
		if w.ShortenAbsoluteLabelsToRelative {
			thisPackage := labelPrefix + filepath.Dir(f.Path)
			// filepath.Dir on Windows uses backslashes as separators, while labels always have slashes.
			if filepath.Separator != '/' {
				thisPackage = strings.Replace(thisPackage, string(filepath.Separator), "/", -1)
			}

			if str.Value == thisPackage {
				str.Value = ":" + path.Base(str.Value)
			} else if strings.HasPrefix(str.Value, thisPackage+":") {
				str.Value = str.Value[len(thisPackage):]
			}
		}

		m := labelRE.FindStringSubmatch(str.Value)
		if m == nil {
			return
		}
		if m[4] != "" && m[4] == m[5] { // e.g. //foo:foo
			str.Value = m[1]
		} else if m[3] != "" && m[4] == "" && m[3] == m[5] { // e.g. @foo//:foo
			str.Value = "@" + m[3]
		}
	}

	// Join and shorten labels within a container of labels (which can be a single
	// label, e.g. a single string expression or a concatenation of them).
	// Gracefully finish if the argument is of a different type.
	fixLabelsWithinAContainer := func(e *Expr) {
		if list, ok := (*e).(*ListExpr); ok {
			for i := range list.List {
				if leaveAlone1(list.List[i]) {
					continue
				}
				joinLabel(&list.List[i])
				shortenLabel(list.List[i])
			}
		}
		if set, ok := (*e).(*SetExpr); ok {
			for i := range set.List {
				if leaveAlone1(set.List[i]) {
					continue
				}
				joinLabel(&set.List[i])
				shortenLabel(set.List[i])
			}
		} else {
			joinLabel(e)
			shortenLabel(*e)
		}
	}

	Walk(f, func(v Expr, stk []Expr) {
		switch v := v.(type) {
		case *CallExpr:
			if leaveAlone(stk, v) {
				return
			}
			for i := range v.List {
				if leaveAlone1(v.List[i]) {
					continue
				}
				as, ok := v.List[i].(*AssignExpr)
				if !ok {
					continue
				}
				key, ok := as.LHS.(*Ident)
				if !ok || !w.IsLabelArg[key.Name] || w.LabelDenyList[callName(v)+"."+key.Name] {
					continue
				}
				if leaveAlone1(as.RHS) {
					continue
				}

				findAndModifyStrings(&as.RHS, fixLabelsWithinAContainer)
			}
		}
	})
}

// callName returns the name of the rule being called by call.
// If the call is not to a literal rule name or a dot expression, callName
// returns "".
func callName(call *CallExpr) string {
	return (&Rule{call, ""}).Kind()
}

// sortCallArgs sorts lists of named arguments to a call.
func sortCallArgs(f *File, w *Rewriter) {

	Walk(f, func(v Expr, stk []Expr) {
		call, ok := v.(*CallExpr)
		if !ok {
			return
		}
		if leaveAlone(stk, call) {
			return
		}
		rule := callName(call)
		if rule == "" {
			rule = "<complex rule kind>"
		}

		// Find the tail of the argument list with named arguments.
		start := len(call.List)
		for start > 0 && argName(call.List[start-1]) != "" {
			start--
		}

		// Record information about each arg into a sortable list.
		var args namedArgs
		for i, x := range call.List[start:] {
			name := argName(x)
			args = append(args, namedArg{ruleNamePriority(w, rule, name), name, i, x})
		}

		// Sort the list and put the args back in the new order.
		if sort.IsSorted(args) {
			return
		}
		sort.Sort(args)
		for i, x := range args {
			call.List[start+i] = x.expr
		}
	})
}

// ruleNamePriority maps a rule argument name to its sorting priority.
// It could use the auto-generated per-rule tables but for now it just
// falls back to the original list.
func ruleNamePriority(w *Rewriter, rule, arg string) int {
	ruleArg := rule + "." + arg
	if val, ok := w.NamePriority[ruleArg]; ok {
		return val
	}
	return w.NamePriority[arg]

	/*
		list := ruleArgOrder[rule]
		if len(list) == 0 {
			return tables.NamePriority[arg]
		}
		for i, x := range list {
			if x == arg {
				return i
			}
		}
		return len(list)
	*/
}

// If x is of the form key=value, argName returns the string key.
// Otherwise argName returns "".
func argName(x Expr) string {
	if as, ok := x.(*AssignExpr); ok {
		if id, ok := as.LHS.(*Ident); ok {
			return id.Name
		}
	}
	return ""
}

// A namedArg records information needed for sorting
// a named call argument into its proper position.
type namedArg struct {
	priority int    // kind of name; first sort key
	name     string // name; second sort key
	index    int    // original index; final sort key
	expr     Expr   // name=value argument
}

// namedArgs is a slice of namedArg that implements sort.Interface
type namedArgs []namedArg

func (x namedArgs) Len() int      { return len(x) }
func (x namedArgs) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x namedArgs) Less(i, j int) bool {
	p := x[i]
	q := x[j]
	if p.priority != q.priority {
		return p.priority < q.priority
	}
	if p.name != q.name {
		return p.name < q.name
	}
	return p.index < q.index
}

// sortStringLists sorts lists of string literals used as specific rule arguments.
func sortStringLists(f *File, w *Rewriter) {
	sortStringList := func(x *Expr) {
		SortStringList(*x)
	}

	Walk(f, func(e Expr, stk []Expr) {
		switch v := e.(type) {
		case *CallExpr:
			if f.Type == TypeDefault || f.Type == TypeBzl {
				// Rule parameters, not applicable to .bzl or default file types
				return
			}
			if leaveAlone(stk, v) {
				return
			}
			rule := callName(v)
			for _, arg := range v.List {
				if leaveAlone1(arg) {
					continue
				}
				as, ok := arg.(*AssignExpr)
				if !ok || leaveAlone1(as) {
					continue
				}
				key, ok := as.LHS.(*Ident)
				if !ok {
					continue
				}
				context := rule + "." + key.Name
				if w.SortableDenylist[context] {
					continue
				}
				if w.IsSortableListArg[key.Name] ||
					w.SortableAllowlist[context] ||
					(!disabled("unsafesort") && allowedSort(context)) {
					if doNotSort(as) {
						deduplicateStringList(as.RHS)
					} else {
						findAndModifyStrings(&as.RHS, sortStringList)
					}
				}
			}
		case *AssignExpr:
			if disabled("unsafesort") {
				return
			}
			// "keep sorted" comment on x = list forces sorting of list.
			if keepSorted(v) {
				findAndModifyStrings(&v.RHS, sortStringList)
			}
		case *KeyValueExpr:
			if disabled("unsafesort") {
				return
			}
			// "keep sorted" before key: list also forces sorting of list.
			if keepSorted(v) {
				findAndModifyStrings(&v.Value, sortStringList)
			}
		case *ListExpr:
			if disabled("unsafesort") {
				return
			}
			// "keep sorted" comment above first list element also forces sorting of list.
			if len(v.List) > 0 && (keepSorted(v) || keepSorted(v.List[0])) {
				findAndModifyStrings(&e, sortStringList)
			}
		}
	})
}

// deduplicateStringList removes duplicates from a list with string expressions
// without reordering its elements.
// Any suffix-comments are lost, any before- and after-comments are preserved.
func deduplicateStringList(x Expr) {
	list, ok := x.(*ListExpr)
	if !ok {
		return
	}

	list.List = deduplicateStringExprs(list.List)
}

// deduplicateStringExprs removes duplicate string expressions from a slice
// without reordering its elements.
// Any suffix-comments are lost, any before- and after-comments are preserved.
func deduplicateStringExprs(list []Expr) []Expr {
	var comments []Comment
	alreadySeen := make(map[string]bool)
	var deduplicated []Expr
	for _, value := range list {
		str, ok := value.(*StringExpr)
		if !ok {
			deduplicated = append(deduplicated, value)
			continue
		}
		strVal := str.Value
		if _, ok := alreadySeen[strVal]; ok {
			// This is a duplicate of a string above.
			// Collect comments so that they're not lost.
			comments = append(comments, str.Comment().Before...)
			comments = append(comments, str.Comment().After...)
			continue
		}
		alreadySeen[strVal] = true
		if len(comments) > 0 {
			comments = append(comments, value.Comment().Before...)
			value.Comment().Before = comments
			comments = nil
		}
		deduplicated = append(deduplicated, value)
	}
	return deduplicated
}

// SortStringList sorts x, a list of strings.
// The list is broken by non-strings and by blank lines and comments into chunks.
// Each chunk is sorted in place.
func SortStringList(x Expr) {
	list, ok := x.(*ListExpr)
	if !ok || len(list.List) < 2 {
		return
	}

	if doNotSort(list.List[0]) {
		list.List = deduplicateStringExprs(list.List)
		return
	}

	forceSort := keepSorted(list) || keepSorted(list.List[0])

	// TODO(bazel-team): Decide how to recognize lists that cannot
	// be sorted. Avoiding all lists with comments avoids sorting
	// lists that say explicitly, in some form or another, why they
	// cannot be sorted. For example, many cc_test rules require
	// certain order in their deps attributes.
	if !forceSort {
		if line, _ := hasComments(list); line {
			deduplicateStringList(list)
			return
		}
	}

	list.List = sortStringExprs(list.List)
}

// findAndModifyStrings finds and modifies string lists with a callback
// function recursively within  the given expression. It doesn't touch all
// string lists it can find, but only top-level lists, lists that are parts of
// concatenated expressions and lists within select statements.
// It calls the callback on the root node and on all relevant inner lists.
// The callback function should gracefully return if called with not appropriate
// arguments.
func findAndModifyStrings(x *Expr, callback func(*Expr)) {
	callback(x)
	switch x := (*x).(type) {
	case *BinaryExpr:
		if x.Op != "+" {
			return
		}
		findAndModifyStrings(&x.X, callback)
		findAndModifyStrings(&x.Y, callback)
	case *CallExpr:
		if ident, ok := x.X.(*Ident); !ok || ident.Name != "select" {
			return
		}
		if len(x.List) == 0 {
			return
		}
		dict, ok := x.List[0].(*DictExpr)
		if !ok {
			return
		}
		for _, kv := range dict.List {
			findAndModifyStrings(&kv.Value, callback)
		}
	}
}

func sortStringExprs(list []Expr) []Expr {
	if len(list) < 2 {
		return list
	}

	// Sort chunks of the list with no intervening blank lines or comments.
	for i := 0; i < len(list); {
		if _, ok := list[i].(*StringExpr); !ok {
			i++
			continue
		}

		j := i + 1
		for ; j < len(list); j++ {
			if str, ok := list[j].(*StringExpr); !ok || len(str.Before) > 0 {
				break
			}
		}

		var chunk []stringSortKey
		for index, x := range list[i:j] {
			chunk = append(chunk, makeSortKey(index, x.(*StringExpr)))
		}
		if !sort.IsSorted(byStringExpr(chunk)) || !isUniq(chunk) {
			before := chunk[0].x.Comment().Before
			chunk[0].x.Comment().Before = nil

			sort.Sort(byStringExpr(chunk))
			chunk = uniq(chunk)

			chunk[0].x.Comment().Before = before
			for offset, key := range chunk {
				list[i+offset] = key.x
			}
			list = append(list[:(i+len(chunk))], list[j:]...)
		}

		i = j
	}

	return list
}

// uniq removes duplicates from a list, which must already be sorted.
// It edits the list in place.
func uniq(sortedList []stringSortKey) []stringSortKey {
	out := sortedList[:0]
	for _, sk := range sortedList {
		if len(out) == 0 || sk.value != out[len(out)-1].value {
			out = append(out, sk)
		}
	}
	return out
}

// isUniq reports whether the sorted list only contains unique elements.
func isUniq(list []stringSortKey) bool {
	for i := range list {
		if i+1 < len(list) && list[i].value == list[i+1].value {
			return false
		}
	}
	return true
}

// If stk describes a call argument like rule(arg=...), callArgName
// returns the name of that argument, formatted as "rule.arg".
func callArgName(stk []Expr) string {
	n := len(stk)
	if n < 2 {
		return ""
	}
	arg := argName(stk[n-1])
	if arg == "" {
		return ""
	}
	call, ok := stk[n-2].(*CallExpr)
	if !ok {
		return ""
	}
	rule, ok := call.X.(*Ident)
	if !ok {
		return ""
	}
	return rule.Name + "." + arg
}

// A stringSortKey records information about a single string literal to be
// sorted. The strings are first grouped into four phases: most strings,
// strings beginning with ":", strings beginning with "//", and strings
// beginning with "@". The next significant part of the comparison is the list
// of elements in the value, where elements are split at `.' and `:'. Finally
// we compare by value and break ties by original index.
type stringSortKey struct {
	phase    int
	split    []string
	value    string
	original int
	x        Expr
}

func makeSortKey(index int, x *StringExpr) stringSortKey {
	key := stringSortKey{
		value:    x.Value,
		original: index,
		x:        x,
	}

	switch {
	case strings.HasPrefix(x.Value, ":"):
		key.phase = 1
	case strings.HasPrefix(x.Value, "//") || (tables.StripLabelLeadingSlashes && !strings.HasPrefix(x.Value, "@")):
		key.phase = 2
	case strings.HasPrefix(x.Value, "@"):
		key.phase = 3
	}

	key.split = strings.Split(strings.Replace(x.Value, ":", ".", -1), ".")
	return key
}

// byStringExpr implements sort.Interface for a list of stringSortKey.
type byStringExpr []stringSortKey

func (x byStringExpr) Len() int      { return len(x) }
func (x byStringExpr) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byStringExpr) Less(i, j int) bool {
	xi := x[i]
	xj := x[j]

	if xi.phase != xj.phase {
		return xi.phase < xj.phase
	}
	for k := 0; k < len(xi.split) && k < len(xj.split); k++ {
		if xi.split[k] != xj.split[k] {
			return xi.split[k] < xj.split[k]
		}
	}
	if len(xi.split) != len(xj.split) {
		return len(xi.split) < len(xj.split)
	}
	if xi.value != xj.value {
		return xi.value < xj.value
	}
	return xi.original < xj.original
}

// fixMultilinePlus turns
//
//	... +
//	[ ... ]
//
//	... +
//	call(...)
//
// into
//
//	... + [
//		...
//	]
//
//	... + call(
//		...
//	)
//
// which typically works better with our aggressively compact formatting.
func fixMultilinePlus(f *File, _ *Rewriter) {

	// List manipulation helpers.
	// As a special case, we treat f([...]) as a list, mainly
	// for glob.

	// isList reports whether x is a list.
	var isList func(x Expr) bool
	isList = func(x Expr) bool {
		switch x := x.(type) {
		case *ListExpr:
			return true
		case *CallExpr:
			if len(x.List) == 1 {
				return isList(x.List[0])
			}
		}
		return false
	}

	// isMultiLine reports whether x is a multiline list.
	var isMultiLine func(Expr) bool
	isMultiLine = func(x Expr) bool {
		switch x := x.(type) {
		case *ListExpr:
			return x.ForceMultiLine || len(x.List) > 1
		case *CallExpr:
			if x.ForceMultiLine || len(x.List) > 1 && !x.ForceCompact {
				return true
			}
			if len(x.List) == 1 {
				return isMultiLine(x.List[0])
			}
		}
		return false
	}

	// forceMultiLine tries to force the list x to use a multiline form.
	// It reports whether it was successful.
	var forceMultiLine func(Expr) bool
	forceMultiLine = func(x Expr) bool {
		switch x := x.(type) {
		case *ListExpr:
			// Already multi line?
			if x.ForceMultiLine {
				return true
			}
			// If this is a list containing a list, force the
			// inner list to be multiline instead.
			if len(x.List) == 1 && forceMultiLine(x.List[0]) {
				return true
			}
			x.ForceMultiLine = true
			return true

		case *CallExpr:
			if len(x.List) == 1 {
				return forceMultiLine(x.List[0])
			}
		}
		return false
	}

	skip := map[Expr]bool{}
	Walk(f, func(v Expr, stk []Expr) {
		if skip[v] {
			return
		}
		bin, ok := v.(*BinaryExpr)
		if !ok || bin.Op != "+" {
			return
		}

		// Found a +.
		// w + x + y + z parses as ((w + x) + y) + z,
		// so chase down the left side to make a list of
		// all the things being added together, separated
		// by the BinaryExprs that join them.
		// Mark them as "skip" so that when Walk recurses
		// into the subexpressions, we won't reprocess them.
		var all []Expr
		for {
			all = append(all, bin.Y, bin)
			bin1, ok := bin.X.(*BinaryExpr)
			if !ok || bin1.Op != "+" {
				break
			}
			bin = bin1
			skip[bin] = true
		}
		all = append(all, bin.X)

		// Because the outermost expression was the
		// rightmost one, the list is backward. Reverse it.
		for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
			all[i], all[j] = all[j], all[i]
		}

		// The 'all' slice is alternating addends and BinaryExpr +'s:
		//	w, +, x, +, y, +, z
		// If there are no lists involved, don't rewrite anything.
		haveList := false
		for i := 0; i < len(all); i += 2 {
			if isList(all[i]) {
				haveList = true
				break
			}
		}
		if !haveList {
			return
		}

		// Okay, there are lists.
		// Consider each + next to a line break.
		for i := 1; i < len(all); i += 2 {
			bin := all[i].(*BinaryExpr)
			if !bin.LineBreak {
				continue
			}

			// We're going to break the line after the +.
			// If it is followed by a list, force that to be
			// multiline instead.
			if forceMultiLine(all[i+1]) {
				bin.LineBreak = false
				continue
			}

			// If the previous list was multiline already,
			// don't bother with the line break after
			// the +.
			if isMultiLine(all[i-1]) {
				bin.LineBreak = false
				continue
			}
		}
	})
}

// isFunctionCall checks whether expr is a call of a function with a given name
func isFunctionCall(expr Expr, name string) bool {
	call, ok := expr.(*CallExpr)
	if !ok {
		return false
	}
	if ident, ok := call.X.(*Ident); ok && ident.Name == name {
		return true
	}
	return false
}

// moveLoadOnTop moves all load statements to the top of the file
func moveLoadOnTop(f *File, _ *Rewriter) {
	if f.Type == TypeWorkspace {
		// Moving load statements in Workspace files can break the semantics
		return
	}
	if isCommentAnywhere(f, "disable=load-on-top") {
		// For backward compatibility. This rewrite used to be a suppressible warning,
		// in some cases it's hard to maintain the position of load statements (e.g.
		// when the file is automatically generated or has automatic transformations
		// applied to it). The rewrite checks for the comment anywhere in the file
		// because it's hard to determine which statement is out of order.
		return
	}

	// Find the misplaced load statements
	misplacedLoads := make(map[int]*LoadStmt)
	firstStmtIndex := -1 // index of the first seen non-load statement
	for i := 0; i < len(f.Stmt); i++ {
		stmt := f.Stmt[i]
		_, isString := stmt.(*StringExpr) // typically a docstring
		_, isComment := stmt.(*CommentBlock)
		isWorkspace := isFunctionCall(stmt, "workspace")
		if isString || isComment || isWorkspace || stmt == nil {
			continue
		}
		load, ok := stmt.(*LoadStmt)
		if !ok {
			if firstStmtIndex == -1 {
				firstStmtIndex = i
			}
			continue
		}
		if firstStmtIndex == -1 {
			continue
		}
		misplacedLoads[i] = load
	}

	// Calculate a fix
	if firstStmtIndex == -1 {
		firstStmtIndex = 0
	}
	offset := len(misplacedLoads)
	if offset == 0 {
		return
	}
	stmtCopy := append([]Expr{}, f.Stmt...)
	for i := range f.Stmt {
		if i < firstStmtIndex {
			// Docstring or a comment in the beginning, skip
			continue
		} else if _, ok := misplacedLoads[i]; ok {
			// A misplaced load statement, should be moved up to the `firstStmtIndex` position
			f.Stmt[firstStmtIndex] = stmtCopy[i]
			firstStmtIndex++
			offset--
			if offset == 0 {
				// No more statements should be moved
				break
			}
		} else {
			// An actual statement (not a docstring or a comment in the beginning), should be moved
			// `offset` positions down.
			f.Stmt[i+offset] = stmtCopy[i]
		}
	}
}

// compressSameOriginLoads merges loads from the same source into one load statement
func compressSameOriginLoads(f *File, _ *Rewriter) {
	// Load statements grouped by their source files
	loads := make(map[string]*LoadStmt)

	for stmtIndex := 0; stmtIndex < len(f.Stmt); stmtIndex++ {
		load, ok := f.Stmt[stmtIndex].(*LoadStmt)
		if !ok {
			continue
		}

		previousLoad, ok := loads[load.Module.Value]
		if !ok {
			loads[load.Module.Value] = load
			continue
		}
		if hasComment(previousLoad, "disable=same-origin-load") ||
			hasComment(load, "disable=same-origin-load") {
			continue
		}

		// Move loaded symbols to the existing load statement
		previousLoad.From = append(previousLoad.From, load.From...)
		previousLoad.To = append(previousLoad.To, load.To...)
		// Remove the current statement
		f.Stmt[stmtIndex] = nil
	}
}

// comparePaths compares two strings as if they were paths (the path delimiter,
// '/', should be treated as smallest symbol).
func comparePaths(path1, path2 string) bool {
	if path1 == path2 {
		return false
	}

	chunks1 := strings.Split(path1, "/")
	chunks2 := strings.Split(path2, "/")

	for i, chunk1 := range chunks1 {
		if i >= len(chunks2) {
			return false
		}
		chunk2 := chunks2[i]
		// Compare case-insensitively
		chunk1Lower := strings.ToLower(chunk1)
		chunk2Lower := strings.ToLower(chunk2)
		if chunk1Lower != chunk2Lower {
			return chunk1Lower < chunk2Lower
		}
	}

	// No case-insensitive difference detected. Likely the difference is just in
	// the case of some symbols, compare case-sensitively for the determinism.
	return path1 <= path2
}

// compareLoadLabels compares two module names
// If one label has explicit repository path (starts with @), it goes first
// If the packages are different, labels are sorted by package name (empty package goes first)
// If the packages are the same, labels are sorted by their name
func compareLoadLabels(load1Label, load2Label string) bool {
	// handle absolute labels with explicit repositories separately to
	// make sure they precede absolute and relative labels without repos
	isExplicitRepo1 := strings.HasPrefix(load1Label, "@")
	isExplicitRepo2 := strings.HasPrefix(load2Label, "@")

	if isExplicitRepo1 != isExplicitRepo2 {
		// Exactly one label has an explicit repository name, it should be the first one.
		return isExplicitRepo1
	}

	// Either both labels have explicit repository names or both don't, compare their packages
	// and break ties using file names if necessary
	module1Parts := strings.SplitN(load1Label, ":", 2)
	package1, filename1 := "", module1Parts[0]
	if len(module1Parts) == 2 {
		package1, filename1 = module1Parts[0], module1Parts[1]
	}
	module2Parts := strings.SplitN(load2Label, ":", 2)
	package2, filename2 := "", module2Parts[0]
	if len(module2Parts) == 2 {
		package2, filename2 = module2Parts[0], module2Parts[1]
	}

	// in case both packages are the same, use file names to break ties
	if package1 == package2 {
		return comparePaths(filename1, filename2)
	}

	// in case one of the packages is empty, the empty one goes first
	if len(package1) == 0 || len(package2) == 0 {
		return len(package1) > 0
	}

	// both packages are non-empty and not equal, so compare them
	return comparePaths(package1, package2)
}

// sortLoadStatements reorders sorts loads lexicographically by the source file,
// but absolute loads have priority over local loads.
func sortLoadStatements(f *File, _ *Rewriter) {
	if isCommentAnywhere(f, "disable=out-of-order-load") {
		// For backward compatibility. This rewrite used to be a suppressible warning,
		// in some cases it's hard to maintain the position of load statements (e.g.
		// when the file is automatically generated or has automatic transformations
		// applied to it). The rewrite checks for the comment anywhere in the file
		// because it's hard to determine which statement is out of order.
		return
	}

	// Consequent chunks of load statements (i.e. without statements of other types between them)
	var loadsChunks [][]*LoadStmt
	newChunk := true // Create a new chunk for the next seen load statement

	for _, stmt := range f.Stmt {
		if stmt == nil {
			// nil statements are no-op and shouldn't break consecutive chains of
			// load statements
			continue
		}
		load, ok := stmt.(*LoadStmt)
		if !ok {
			newChunk = true
			continue
		}
		if newChunk {
			// There's a non-load statement between this load and the previous load
			loadsChunks = append(loadsChunks, []*LoadStmt{})
			newChunk = false
		}
		loadsChunks[len(loadsChunks)-1] = append(loadsChunks[len(loadsChunks)-1], load)
	}

	// Sort and flatten the chunks
	var sortedLoads []*LoadStmt
	for _, chunk := range loadsChunks {
		sortedChunk := append([]*LoadStmt{}, chunk...)

		sort.SliceStable(sortedChunk, func(i, j int) bool {
			load1Label := (sortedChunk)[i].Module.Value
			load2Label := (sortedChunk)[j].Module.Value
			return compareLoadLabels(load1Label, load2Label)
		})
		sortedLoads = append(sortedLoads, sortedChunk...)
	}

	// Calculate the replacements
	loadIndex := 0
	for stmtIndex := range f.Stmt {
		if _, ok := f.Stmt[stmtIndex].(*LoadStmt); !ok {
			continue
		}
		f.Stmt[stmtIndex] = sortedLoads[loadIndex]
		loadIndex++
	}
}

// sortAllLoadArgs sorts all load arguments in the file
func sortAllLoadArgs(f *File, _ *Rewriter) {
	Walk(f, func(v Expr, stk []Expr) {
		if load, ok := v.(*LoadStmt); ok {
			SortLoadArgs(load)
		}
	})
}

func sortUseRepoPositionals(f *File, _ *Rewriter) {
	Walk(f, func(v Expr, stk []Expr) {
		if call, ok := v.(*CallExpr); ok {
			// The first argument of a valid use_repo call is always a module extension proxy, so we
			// do not need to sort calls with less than three arguments.
			if ident, ok := call.X.(*Ident); !ok || ident.Name != "use_repo" || len(call.List) < 3 {
				return
			}
			// Respect the "do not sort" comment on both the first argument and the first repository
			// name.
			if doNotSort(call) || doNotSort(call.List[0]) || doNotSort(call.List[1]) {
				call.List = deduplicateStringExprs(call.List)
			} else {
				// Keyword arguments do not have to be sorted here as this has already been done by
				// the generic callsort rewriter pass.
				call.List = sortStringExprs(call.List)
			}
		}
	})
}

// hasComments reports whether any comments are associated with
// the list or its elements.
func hasComments(list *ListExpr) (line, suffix bool) {
	com := list.Comment()
	if len(com.Before) > 0 || len(com.After) > 0 || len(list.End.Before) > 0 {
		line = true
	}
	if len(com.Suffix) > 0 {
		suffix = true
	}
	for _, elem := range list.List {
		com := elem.Comment()
		if len(com.Before) > 0 {
			line = true
		}
		if len(com.Suffix) > 0 {
			suffix = true
		}
	}
	return
}

// A wrapper for a LoadStmt's From and To slices for consistent sorting of their contents.
// It's assumed that the following slices have the same length. The contents are sorted by
// the `To` attribute, but all items with equal "From" and "To" parts are placed before the items
// with different parts.
type loadArgs struct {
	From     []*Ident
	To       []*Ident
	modified bool
}

func (args *loadArgs) Len() int {
	return len(args.From)
}

func (args *loadArgs) Swap(i, j int) {
	args.From[i], args.From[j] = args.From[j], args.From[i]
	args.To[i], args.To[j] = args.To[j], args.To[i]
	args.modified = true
}

func (args *loadArgs) Less(i, j int) bool {
	// Arguments with equal "from" and "to" parts are prioritized
	equalI := args.From[i].Name == args.To[i].Name
	equalJ := args.From[j].Name == args.To[j].Name
	if equalI != equalJ {
		// If equalI and !equalJ, return true, otherwise false.
		// Equivalently, return equalI.
		return equalI
	}
	return args.To[i].Name < args.To[j].Name
}

// deduplicate assumes that the args are sorted and prioritize arguments that
// are defined later (e.g. `a = "x", a = "y"` is shortened to `a = "y"`).
func (args *loadArgs) deduplicate() {
	var from, to []*Ident
	for i := range args.To {
		if len(to) > 0 && to[len(to)-1].Name == args.To[i].Name {
			// Overwrite
			from[len(from)-1] = args.From[i]
			to[len(to)-1] = args.To[i]
			args.modified = true
		} else {
			// Just append
			from = append(from, args.From[i])
			to = append(to, args.To[i])
		}
	}
	args.From = from
	args.To = to
}

// SortLoadArgs sorts a load statement arguments (lexicographically, but positional first)
func SortLoadArgs(load *LoadStmt) bool {
	args := loadArgs{From: load.From, To: load.To}
	sort.Sort(&args)
	args.deduplicate()
	load.From = args.From
	load.To = args.To
	return args.modified
}

// formatDocstrings fixes the indentation and trailing whitespace of docstrings
func formatDocstrings(f *File, _ *Rewriter) {
	Walk(f, func(v Expr, stk []Expr) {
		def, ok := v.(*DefStmt)
		if !ok || len(def.Body) == 0 {
			return
		}
		docstring, ok := def.Body[0].(*StringExpr)
		if !ok || !docstring.TripleQuote {
			return
		}

		oldIndentation := docstring.Start.LineRune - 1 // LineRune starts with 1
		newIndentation := nestedIndentation * len(stk)

		// Operate on Token, not Value, because their line breaks can be different if a line ends with
		// a backslash.
		updatedToken := formatString(docstring.Token, oldIndentation, newIndentation)
		if updatedToken != docstring.Token {
			docstring.Token = updatedToken
			// Update the value to keep it consistent with Token
			docstring.Value, _, _ = Unquote(updatedToken)
		}
	})
}

// formatString modifies a string value of a docstring to match the new indentation level and
// to remove trailing whitespace from its lines.
func formatString(value string, oldIndentation, newIndentation int) string {
	difference := newIndentation - oldIndentation
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if i == 0 {
			// The first line shouldn't be touched because it starts right after ''' or """
			continue
		}
		if difference > 0 {
			line = strings.Repeat(" ", difference) + line
		} else {
			for i, rune := range line {
				if i == -difference || rune != ' ' {
					line = line[i:]
					break
				}
			}
		}
		if i != len(lines)-1 {
			// Remove trailing space from the line unless it's the last line that's responsible
			// for the indentation of the closing `"""`
			line = strings.TrimRight(line, " ")
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// argumentType returns an integer by which funcall arguments can be sorted:
// 1 for positional, 2 for named, 3 for *args, 4 for **kwargs
func argumentType(expr Expr) int {
	switch expr := expr.(type) {
	case *UnaryExpr:
		switch expr.Op {
		case "**":
			return 4
		case "*":
			return 3
		}
	case *AssignExpr:
		return 2
	}
	return 1
}

// reorderArguments fixes the order of arguments of a function call
// (positional, named, *args, **kwargs)
func reorderArguments(f *File, _ *Rewriter) {
	Walk(f, func(expr Expr, stack []Expr) {
		call, ok := expr.(*CallExpr)
		if !ok {
			return
		}
		compare := func(i, j int) bool {
			// Keep nil nodes at their place. They are no-op for the formatter but can
			// be useful for the linter which expects them to not move.
			if call.List[i] == nil || call.List[j] == nil {
				return false
			}
			return argumentType(call.List[i]) < argumentType(call.List[j])
		}
		if !sort.SliceIsSorted(call.List, compare) {
			sort.SliceStable(call.List, compare)
		}
	})
}

// editOctals inserts 'o' into octal numbers to make it more obvious they are octal
// 0123 -> 0o123
func editOctals(f *File, _ *Rewriter) {
	Walk(f, func(expr Expr, stack []Expr) {
		l, ok := expr.(*LiteralExpr)
		if !ok {
			return
		}
		if len(l.Token) > 1 && l.Token[0] == '0' && l.Token[1] >= '0' && l.Token[1] <= '9' {
			l.Token = "0o" + l.Token[1:]
		}
	})
}

// removeParens removes trivial parens
func removeParens(f *File, _ *Rewriter) {
	var simplify func(expr Expr, stack []Expr) Expr
	simplify = func(expr Expr, stack []Expr) Expr {
		// Look for parenthesized expressions, ignoring those with
		// comments and those that are intentionally multiline.
		pa, ok := expr.(*ParenExpr)
		if !ok || pa.ForceMultiLine {
			return expr
		}
		if len(pa.Comment().Before) > 0 || len(pa.Comment().After) > 0 || len(pa.Comment().Suffix) > 0 {
			return expr
		}

		switch x := pa.X.(type) {
		case *Comprehension, *DictExpr, *Ident, *ListExpr, *LiteralExpr, *ParenExpr, *SetExpr, *StringExpr:
			// These expressions don't need parens, remove them (recursively).
			return Edit(x, simplify)
		case *CallExpr:
			// Parens might be needed if the callable is multiline.
			start, end := x.X.Span()
			if start.Line == end.Line {
				return Edit(x, simplify)
			}
		}
		return expr
	}

	Edit(f, simplify)
}
