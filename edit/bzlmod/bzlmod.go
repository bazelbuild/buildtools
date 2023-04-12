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
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

// Proxies returns the names of extension proxies (i.e. the names of variables to which the result
// of a use_extension call is assigned) for the given extension with the given value of the
// dev_dependency attribute.
func Proxies(f *build.File, rawExtBzlFile string, extName string, dev bool) []string {
	apparentModuleName := getApparentModuleName(f)
	extBzlFile := normalizeLabelString(rawExtBzlFile, apparentModuleName)

	var proxies []string
	for _, stmt := range f.Stmt {
		proxy, rawBzlFile, name, isDev := parseUseExtension(stmt)
		if proxy == "" || isDev != dev {
			continue
		}
		bzlFile := normalizeLabelString(rawBzlFile, apparentModuleName)
		if bzlFile == extBzlFile && name == extName {
			proxies = append(proxies, proxy)
		}
	}

	return proxies
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

// AddRepoUsages adds the given repos to the given use_repo calls without introducing duplicate
// arguments.
// useRepos must not be empty.
// Keyword arguments are preserved but adding them is currently not supported.
func AddRepoUsages(useRepos []*build.CallExpr, repos ...string) {
	if len(repos) == 0 {
		return
	}
	if len(useRepos) == 0 {
		panic("useRepos must not be empty")
	}

	seen := make(map[string]struct{})
	for _, useRepo := range useRepos {
		if len(useRepo.List) == 0 {
			// Invalid use_repo call, skip.
			continue
		}
		for _, arg := range useRepo.List[1:] {
			seen[repoFromUseRepoArg(arg)] = struct{}{}
		}
	}

	lastUseRepo := getLastUseRepo(useRepos)
	for _, repo := range repos {
		if _, ok := seen[repo]; ok {
			continue
		}
		// Sorting of use_repo arguments is handled by Buildify.
		// TODO: Add a keyword argument instead if repo is of the form "key=value".
		lastUseRepo.List = append(lastUseRepo.List, &build.StringExpr{Value: repo})
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
	// This implements
	// https://github.com/bazelbuild/bazel/blob/dd822392db96bb7bccdb673414a20c4b91e3dbc1/src/main/java/com/google/devtools/build/lib/bazel/bzlmod/ModuleFileGlobals.java#L416
	// with the assumption that the current module is the root module.
	if strings.HasPrefix(rawLabel, "//") {
		// Relative labels always refer to the current module.
		return "@" + apparentModuleName + rawLabel
	} else if strings.HasPrefix(rawLabel, "@//") {
		// In the root module only, this syntax refer to the module. Since we are inspecting its
		// module file as a tool, we can assume that the current module is the root module.
		return "@" + apparentModuleName + rawLabel[1:]
	} else {
		return rawLabel
	}
}

func parseUseExtension(stmt build.Expr) (proxy string, bzlFile string, name string, dev bool) {
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
			keywordArg, ok := arg.(*build.AssignExpr)
			if !ok {
				continue
			}
			argName, ok := keywordArg.LHS.(*build.Ident)
			if !ok || argName.Name != "dev_dependency" {
				continue
			}
			argValue, ok := keywordArg.RHS.(*build.Ident)
			// We assume that any expression other than "False" evaluates to
			// True as otherwise there would be no reason to specify the
			// argument - MODULE.bazel files are entirely static, so every
			// expression always evaluates to the same value.
			if !ok || argValue.Name != "False" {
				dev = true
				break
			}
		}
	}
	return assign.LHS.(*build.Ident).Name, bzlFileExpr.Value, nameExpr.Value, dev
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
func lastProxyUsage(f *build.File, proxies []string) (lastUsage int, proxy string) {
	proxiesSet := make(map[string]struct{})
	for _, p := range proxies {
		proxiesSet[p] = struct{}{}
	}

	lastUsage = -1
	for i, stmt := range f.Stmt {
		proxy, _, _, _ = parseUseExtension(stmt)
		if proxy != "" {
			_, isUsage := proxiesSet[proxy]
			if isUsage {
				lastUsage = i
				continue
			}
		}

		proxy = parseTag(stmt)
		if proxy != "" {
			_, isUsage := proxiesSet[proxy]
			if isUsage {
				lastUsage = i
				continue
			}
		}
	}

	return lastUsage, proxy
}
