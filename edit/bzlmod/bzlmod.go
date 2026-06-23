/*
Copyright 2023 Google LLC

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

// Package bzlmod contains functions for working with MODULE.bazel files.
package bzlmod

import (
	"path"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/labels"
)

// Proxies returns the names of extension proxies (i.e. the names of variables to which the result
// of a use_extension call is assigned) for the given extension with the given value of the
// dev_dependency attribute.
// Extension proxies created with "isolate = True" are ignored.
func Proxies(f *build.File, rawExtBzlFile string, extName string, dev bool) []string {
	apparentModuleName := getApparentModuleName(f)
	extBzlFile := normalizeLabelString(rawExtBzlFile, apparentModuleName)

	var proxies []string
	for _, stmt := range f.Stmt {
		proxy, rawBzlFile, name, isDev, isIsolated := parseUseExtension(stmt)
		if proxy == "" || isDev != dev || isIsolated {
			continue
		}
		bzlFile := normalizeLabelString(rawBzlFile, apparentModuleName)
		if bzlFile == extBzlFile && name == extName {
			proxies = append(proxies, proxy)
		}
	}

	return proxies
}

// AllProxies returns the names of all extension proxies (i.e. the names of variables to which the
// result of a use_extension call is assigned) corresponding to the same extension usage as the
// given proxy.
// For an isolated extension usage, a list containing only the given proxy is returned.
// For a non-isolated extension usage, the proxies of all non-isolated extension usages of the same
// extension with the same value for the dev_dependency parameter are returned.
// If the given proxy is not an extension proxy, nil is returned.
func AllProxies(f *build.File, proxy string) []string {
	for _, stmt := range f.Stmt {
		proxyCandidate, rawBzlFile, name, isDev, isIsolated := parseUseExtension(stmt)
		if proxyCandidate == proxy {
			if isIsolated {
				return []string{proxy}
			}
			return Proxies(f, rawBzlFile, name, isDev)
		}
	}
	return nil
}

// UseRepos returns the use_repo calls that use the given proxies.
func UseRepos(f *build.File, proxies []string) []*build.CallExpr {
	proxiesSet := make(map[string]struct{})
	for _, p := range proxies {
		proxiesSet[p] = struct{}{}
	}

	var useRepos []*build.CallExpr
	for _, stmt := range f.Stmt {
		if _, ok := stmt.(*build.CallExpr); !ok {
			continue
		}
		call := stmt.(*build.CallExpr)
		if _, ok := call.X.(*build.Ident); !ok {
			continue
		}
		if call.X.(*build.Ident).Name != "use_repo" || len(call.List) < 1 {
			continue
		}
		proxy, ok := call.List[0].(*build.Ident)
		if !ok {
			continue
		}
		if _, ok := proxiesSet[proxy.Name]; !ok {
			continue
		}
		useRepos = append(useRepos, call)
	}

	return useRepos
}

// NewUseRepo inserts and returns a new use_repo call after the last usage of any of the given
// proxies, where a usage is either a use_extension call or a tag definition.
func NewUseRepo(f *build.File, proxies []string) (*build.File, *build.CallExpr) {
	lastUsage, proxy := lastProxyUsage(f, proxies)
	if lastUsage == -1 {
		return f, nil
	}

	useRepo := &build.CallExpr{
		X: &build.Ident{Name: "use_repo"},
		List: []build.Expr{
			&build.Ident{Name: proxy},
		},
	}
	stmt := append(f.Stmt[:lastUsage+1], append([]build.Expr{useRepo}, f.Stmt[lastUsage+1:]...)...)

	return &build.File{Path: f.Path, Comments: f.Comments, Stmt: stmt, Type: build.TypeModule}, useRepo
}

// repoToAdd represents a repository to be added to a use_repo call.
type repoToAdd struct {
	key          string
	value        string
	isMapping    bool
	isValidIdent bool
}

// AddRepoUsages adds the given repos to the given use_repo calls without introducing duplicate
// arguments.
// useRepos must not be empty.
// Keyword arguments are preserved and can be added using the syntax "key=value".
// For invalid identifiers (e.g., containing dots), dict unpacking syntax is used.
// If a mapping conflicts with an existing one (same key or same value), the existing one is replaced.
func AddRepoUsages(useRepos []*build.CallExpr, repos ...string) {
	if len(repos) == 0 {
		return
	}
	if len(useRepos) == 0 {
		panic("useRepos must not be empty")
	}

	// Parse all repos to add and track which keys/values are being added
	var reposToAdd []repoToAdd
	mappingValuesToAdd := make(map[string]struct{}) // Values being added as mappings
	keysToAdd := make(map[string]struct{})

	for _, repo := range repos {
		key, value, isMapping := parseRepoMapping(repo)

		// Simplify foo=foo to just foo (positional argument)
		if isMapping && key == value {
			isMapping = false
		}

		validIdent := isValidIdentifier(key)

		reposToAdd = append(reposToAdd, repoToAdd{
			key:          key,
			value:        value,
			isMapping:    isMapping,
			isValidIdent: validIdent,
		})

		if isMapping {
			mappingValuesToAdd[value] = struct{}{}
			keysToAdd[key] = struct{}{}
		}
	}

	// Track existing values to detect duplicates when adding positional args
	existingValues := collectExistingValues(useRepos)

	// Remove conflicting arguments from all use_repo calls
	removeConflictingRepos(useRepos, mappingValuesToAdd, keysToAdd)

	// Add new repos to the last use_repo call
	lastUseRepo := getLastUseRepo(useRepos)
	addNewRepos(lastUseRepo, reposToAdd, existingValues)
}

// collectExistingValues returns a set of all repository values currently referenced in use_repo calls.
func collectExistingValues(useRepos []*build.CallExpr) map[string]struct{} {
	existingValues := make(map[string]struct{})
	for _, useRepo := range useRepos {
		if len(useRepo.List) == 0 {
			continue
		}
		for _, arg := range useRepo.List[1:] {
			if val := repoFromUseRepoArg(arg); val != "" {
				existingValues[val] = struct{}{}
			}
		}
	}
	return existingValues
}

// removeConflictingRepos removes arguments from use_repo calls that conflict with repos being added.
func removeConflictingRepos(useRepos []*build.CallExpr, mappingValuesToAdd, keysToAdd map[string]struct{}) {
	for _, useRepo := range useRepos {
		if len(useRepo.List) == 0 {
			continue
		}

		var newArgs []build.Expr
		var dictToUpdate *build.DictExpr

		for _, arg := range useRepo.List[1:] {
			shouldKeep := true

			switch arg := arg.(type) {
			case *build.StringExpr:
				// Positional argument: only remove if a mapping is being added with this value
				if _, conflicts := mappingValuesToAdd[arg.Value]; conflicts {
					shouldKeep = false
				}
			case *build.AssignExpr:
				// Keyword argument: remove if same key is being added
				if ident, ok := arg.LHS.(*build.Ident); ok {
					if _, conflicts := keysToAdd[ident.Name]; conflicts {
						shouldKeep = false
					}
				}
				// Or if same value is being added as a mapping (replacement)
				if str, ok := arg.RHS.(*build.StringExpr); ok {
					if _, conflicts := mappingValuesToAdd[str.Value]; conflicts {
						shouldKeep = false
					}
				}
			case *build.UnaryExpr:
				// Dict unpacking: we'll handle this separately
				if arg.Op == "**" {
					if dict, ok := arg.X.(*build.DictExpr); ok {
						dictToUpdate = dict
						shouldKeep = false // Remove from list, we'll re-add after filtering
					}
				}
			}

			if shouldKeep {
				newArgs = append(newArgs, arg)
			}
		}

		// Filter dict entries if we have a dict to update
		if dictToUpdate != nil {
			filteredDictEntries := filterDictEntries(dictToUpdate, mappingValuesToAdd, keysToAdd)
			// Only keep dict if it still has entries after filtering
			if len(filteredDictEntries) > 0 {
				dictToUpdate.List = filteredDictEntries
				newArgs = append(newArgs, &build.UnaryExpr{Op: "**", X: dictToUpdate})
			}
		}

		useRepo.List = append(useRepo.List[:1], newArgs...)
	}
}

// filterDictEntries removes dict entries that conflict with repos being added.
// Filters in-place to reuse the underlying array for better memory efficiency.
func filterDictEntries(dict *build.DictExpr, mappingValuesToAdd, keysToAdd map[string]struct{}) []*build.KeyValueExpr {
	n := 0
	for _, kv := range dict.List {
		shouldKeep := true
		// Check if key conflicts
		if keyStr, ok := kv.Key.(*build.StringExpr); ok {
			if _, conflicts := keysToAdd[keyStr.Value]; conflicts {
				shouldKeep = false
			}
		}
		// Check if value conflicts with a mapping being added
		if valStr, ok := kv.Value.(*build.StringExpr); ok {
			if _, conflicts := mappingValuesToAdd[valStr.Value]; conflicts {
				shouldKeep = false
			}
		}
		if shouldKeep {
			dict.List[n] = kv
			n++
		}
	}
	return dict.List[:n]
}

// addNewRepos adds new repository arguments to the given use_repo call.
func addNewRepos(useRepo *build.CallExpr, reposToAdd []repoToAdd, existingValues map[string]struct{}) {
	var dictEntries []*build.KeyValueExpr

	for _, repo := range reposToAdd {
		if !repo.isMapping {
			// Simple positional argument: use_repo(ext, "foo")
			// Skip if already exists (to preserve existing mappings)
			if _, exists := existingValues[repo.value]; !exists {
				useRepo.List = append(useRepo.List, &build.StringExpr{Value: repo.value})
			}
		} else if repo.isValidIdent {
			// Valid identifier as keyword argument: use_repo(ext, bar = "baz")
			useRepo.List = append(useRepo.List, &build.AssignExpr{
				LHS: &build.Ident{Name: repo.key},
				Op:  "=",
				RHS: &build.StringExpr{Value: repo.value},
			})
		} else {
			// Invalid identifier, need dict unpacking: use_repo(ext, **{"foo.2": "foo"})
			dictEntries = append(dictEntries, &build.KeyValueExpr{
				Key:   &build.StringExpr{Value: repo.key},
				Value: &build.StringExpr{Value: repo.value},
			})
		}
	}

	// If we have dict entries, add or extend the **kwargs dict
	if len(dictEntries) > 0 {
		addOrExtendDictUnpack(useRepo, dictEntries)
	}
}

// RemoveRepoUsages removes the given repos from the given use_repo calls.
// Repositories are identified via their names as exported by the module extension (i.e. the value
// rather than the key in the case of keyword arguments).
func RemoveRepoUsages(useRepos []*build.CallExpr, repos ...string) {
	if len(useRepos) == 0 || len(repos) == 0 {
		return
	}

	toRemove := make(map[string]struct{})
	for _, repo := range repos {
		toRemove[repo] = struct{}{}
	}

	for _, useRepo := range useRepos {
		if len(useRepo.List) == 0 {
			// Invalid use_repo call, skip.
			continue
		}
		var args []build.Expr
		// Skip over ext in use_repo(ext, ...).
		for _, arg := range useRepo.List[1:] {
			repo := repoFromUseRepoArg(arg)
			if _, remove := toRemove[repo]; !remove {
				args = append(args, arg)
			}
		}
		useRepo.List = append(useRepo.List[:1], args...)
	}
}

func getLastUseRepo(useRepos []*build.CallExpr) *build.CallExpr {
	var lastUseRepo *build.CallExpr
	for _, useRepo := range useRepos {
		if lastUseRepo == nil || useRepo.Pos.Byte > lastUseRepo.Pos.Byte {
			lastUseRepo = useRepo
		}
	}
	return lastUseRepo
}

// repoFromUseRepoArg returns the repository name used by the module extension itself from a
// use_repo argument.
func repoFromUseRepoArg(arg build.Expr) string {
	switch arg := arg.(type) {
	case *build.StringExpr:
		// use_repo(ext, "repo") --> repo
		return arg.Value
	case *build.AssignExpr:
		// use_repo(ext, my_repo = "repo") --> repo
		if repo, ok := arg.RHS.(*build.StringExpr); ok {
			return repo.Value
		}
	}
	return ""
}

// getApparentModuleName returns the apparent name used for the repository of the module defined
// in the given MODULE.bazel file.
func getApparentModuleName(f *build.File) string {
	apparentName := ""

	for _, module := range f.Rules("module") {
		if repoName := module.AttrString("repo_name"); repoName != "" {
			apparentName = repoName
		} else if name := module.AttrString("name"); name != "" {
			apparentName = name
		}
	}

	return apparentName
}

// normalizeLabelString converts a label string into the form @apparent_name//path/to:target.
func normalizeLabelString(rawLabel, apparentModuleName string) string {
	label := labels.ParseRelative(rawLabel, "")
	if label.Repository == "" {
		// This branch is taken in two different cases:
		// 1. The label is relative. In this case, labels.ParseRelative populates the Package field
		//    but not the Repository field.
		// 2. The label is of the form "@//pkg:extension.bzl". Normalize to spelling out the
		//    apparent name of the root module. Note that this syntax is only allowed in the root
		//    module, but since we are inspecting its module file as a tool, we can assume that the
		//    current module is the root module.
		label.Repository = apparentModuleName
	}
	return label.Format()
}

func parseUseExtension(stmt build.Expr) (proxy string, bzlFile string, name string, dev bool, isolate bool) {
	assign, ok := stmt.(*build.AssignExpr)
	if !ok {
		return
	}
	if _, ok = assign.LHS.(*build.Ident); !ok {
		return
	}
	if _, ok = assign.RHS.(*build.CallExpr); !ok {
		return
	}
	call := assign.RHS.(*build.CallExpr)
	if _, ok = call.X.(*build.Ident); !ok {
		return
	}
	if call.X.(*build.Ident).Name != "use_extension" {
		return
	}
	if len(call.List) < 2 {
		// Missing required positional arguments.
		return
	}
	bzlFileExpr, ok := call.List[0].(*build.StringExpr)
	if !ok {
		return
	}
	nameExpr, ok := call.List[1].(*build.StringExpr)
	if !ok {
		return
	}
	// Check for the optional dev_dependency keyword argument.
	if len(call.List) > 2 {
		for _, arg := range call.List[2:] {
			dev = dev || parseBooleanKeywordArg(arg, "dev_dependency")
			isolate = isolate || parseBooleanKeywordArg(arg, "isolate")
		}
	}
	return assign.LHS.(*build.Ident).Name, bzlFileExpr.Value, nameExpr.Value, dev, isolate
}

// parseBooleanKeywordArg parses a keyword argument of type bool that is assumed to default to
// False.
func parseBooleanKeywordArg(arg build.Expr, name string) bool {
	keywordArg, ok := arg.(*build.AssignExpr)
	if !ok {
		return false
	}
	argName, ok := keywordArg.LHS.(*build.Ident)
	if !ok || argName.Name != name {
		return false
	}
	argValue, ok := keywordArg.RHS.(*build.Ident)
	// We assume that any expression other than "False" evaluates to True as otherwise there would
	// be no reason to specify the argument - MODULE.bazel files are entirely static with no
	// external inputs, so every expression always evaluates to the same value.
	if ok && argValue.Name == "False" {
		return false
	}
	return true
}

func parseTag(stmt build.Expr) string {
	call, ok := stmt.(*build.CallExpr)
	if !ok {
		return ""
	}
	if _, ok := call.X.(*build.DotExpr); !ok {
		return ""
	}
	dot := call.X.(*build.DotExpr)
	if _, ok := dot.X.(*build.Ident); !ok {
		return ""
	}
	return dot.X.(*build.Ident).Name
}

// lastProxyUsage returns the index of the last statement in the given file that uses one of the
// given extension proxies (either in a use_extension assignment or tag call). If no such statement
// exists, -1 is returned.
func lastProxyUsage(f *build.File, proxies []string) (lastUsage int, lastProxy string) {
	proxiesSet := make(map[string]struct{})
	for _, p := range proxies {
		proxiesSet[p] = struct{}{}
	}

	lastUsage = -1
	for i, stmt := range f.Stmt {
		proxy, _, _, _, _ := parseUseExtension(stmt)
		if proxy != "" {
			_, isUsage := proxiesSet[proxy]
			if isUsage {
				lastUsage = i
				lastProxy = proxy
				continue
			}
		}

		proxy = parseTag(stmt)
		if proxy != "" {
			_, isUsage := proxiesSet[proxy]
			if isUsage {
				lastUsage = i
				lastProxy = proxy
				continue
			}
		}
	}

	return lastUsage, lastProxy
}

// ExtractModuleToApparentNameMapping collects the mapping of module names (e.g. "rules_go") to
// user-configured apparent names (e.g. "my_rules_go") from the repo's MODULE.bazel, if it exists.
// The given function is called with a repo-relative, slash-separated path and should return the
// content of the MODULE.bazel or *.MODULE.bazel file at that path, or nil if the file does not
// exist.
// See https://bazel.build/external/module#repository_names_and_strict_deps for more information on
// apparent names.
func ExtractModuleToApparentNameMapping(fileReader func(relPath string) *build.File) func(string) string {
	moduleToApparentName := collectApparentNames(fileReader, "MODULE.bazel")

	return func(moduleName string) string {
		return moduleToApparentName[moduleName]
	}
}

// Collects the mapping of module names (e.g. "rules_go") to user-configured apparent names (e.g.
// "my_rules_go"). See https://bazel.build/external/module#repository_names_and_strict_deps for more
// information on apparent names.
func collectApparentNames(fileReader func(relPath string) *build.File, relPath string) map[string]string {
	apparentNames := make(map[string]string)
	seenFiles := make(map[string]struct{})
	filesToProcess := []string{relPath}

	for len(filesToProcess) > 0 {
		f := filesToProcess[0]
		filesToProcess = filesToProcess[1:]
		if _, seen := seenFiles[f]; seen {
			continue
		}
		seenFiles[f] = struct{}{}
		bf := fileReader(f)
		if bf == nil {
			return nil
		}
		names, includeLabels := collectApparentNamesAndIncludes(bf)
		for name, apparentName := range names {
			apparentNames[name] = apparentName
		}
		for _, includeLabel := range includeLabels {
			l := labels.Parse(includeLabel)
			p := path.Join(l.Package, l.Target)
			filesToProcess = append(filesToProcess, p)
		}
	}

	return apparentNames
}

func collectApparentNamesAndIncludes(f *build.File) (map[string]string, []string) {
	apparentNames := make(map[string]string)
	var includeLabels []string

	for _, dep := range f.Rules("") {
		if dep.ExplicitName() == "" {
			if ident, ok := dep.Call.X.(*build.Ident); !ok || ident.Name != "include" {
				continue
			}
			if len(dep.Call.List) != 1 {
				continue
			}
			if str, ok := dep.Call.List[0].(*build.StringExpr); ok {
				includeLabels = append(includeLabels, str.Value)
			}
			continue
		}
		if dep.Kind() != "module" && dep.Kind() != "bazel_dep" {
			continue
		}
		// We support module in addition to bazel_dep to handle language repos that use Gazelle to
		// manage their own BUILD files.
		if name := dep.AttrString("name"); name != "" {
			if repoName := dep.AttrString("repo_name"); repoName != "" {
				apparentNames[name] = repoName
			} else {
				apparentNames[name] = name
			}
		}
	}

	return apparentNames, includeLabels
}

// parseRepoMapping parses a repo string which may be in the form "key=value" or just "value".
// Returns (key, value, isMapping) where isMapping indicates if it was a mapping.
func parseRepoMapping(repo string) (key, value string, isMapping bool) {
	key, value, found := strings.Cut(repo, "=")
	if found {
		return key, value, true
	}
	return "", repo, false
}

// starlarkReservedKeywords contains Starlark reserved keywords that cannot be used as identifiers.
var starlarkReservedKeywords = map[string]bool{
	"and":      true,
	"break":    true,
	"continue": true,
	"def":      true,
	"elif":     true,
	"else":     true,
	"for":      true,
	"if":       true,
	"in":       true,
	"lambda":   true,
	"load":     true,
	"not":      true,
	"or":       true,
	"pass":     true,
	"return":   true,
	"while":    true,
	"False":    true,
	"None":     true,
	"True":     true,
}

// isValidIdentifier checks if a string is a valid Starlark identifier.
// Per Starlark rules: must start with ASCII letter (a-z, A-Z) or underscore,
// followed by ASCII letters, digits (0-9), or underscores, and must not be a reserved keyword.
// This matches Bazel's Lexer.scanIdentifier behavior.
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Check if it's a reserved keyword
	if starlarkReservedKeywords[s] {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 0 {
			// First character must be ASCII letter or underscore
			if !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			// Subsequent characters must be ASCII letter, digit, or underscore
			if !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_') {
				return false
			}
		}
	}

	return true
}

// addOrExtendDictUnpack adds dict entries to an existing **kwargs dict unpacking expression,
// or creates a new one if none exists.
func addOrExtendDictUnpack(useRepo *build.CallExpr, newEntries []*build.KeyValueExpr) {
	// Look for an existing **dict unpacking expression
	for _, arg := range useRepo.List {
		if unary, ok := arg.(*build.UnaryExpr); ok && unary.Op == "**" {
			if dict, ok := unary.X.(*build.DictExpr); ok {
				// Found existing dict unpacking, add new entries to it
				dict.List = append(dict.List, newEntries...)
				return
			}
		}
	}

	// No existing **dict found, create a new one
	useRepo.List = append(useRepo.List, &build.UnaryExpr{
		Op: "**",
		X: &build.DictExpr{
			List: newEntries,
		},
	})
}
