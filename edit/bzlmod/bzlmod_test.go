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

package bzlmod

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/bazelbuild/buildtools/build"
)

const proxiesNoModuleHeader = ``

const proxiesModuleNameHeader = `
module(name = "name")
`

const proxiesModuleRepoNameHeader = `
module(name = "name", repo_name = "repo_name")
`

const proxiesBody = `
prox1 = use_extension("@name//bzl:extensions.bzl", "ext")
prox1.use(name = "foo")
isolated_prox1 = use_extension("@name//bzl:extensions.bzl", "ext", dev_dependency = False, isolate = True)
isolated_prox1.use(name = "foo")
prox2 = use_extension("@name//bzl:extensions.bzl", "ext", dev_dependency = True)
prox2.use(list = ["foo", "bar"])
# some comment
prox3 = use_extension("@repo_name//bzl:extensions.bzl", "ext")
prox3.use(label = "@dep//:bar")
isolated_prox2 = use_extension("@name//bzl:extensions.bzl", "ext", dev_dependency = True, isolate = True)
isolated_prox2.use(name = "foo")
prox4 = use_extension("@repo_name//bzl:extensions.bzl", "ext", dev_dependency = True)
prox4.use(dict = {"foo": "bar"})
isolated_prox3 = use_extension("@name//bzl:extensions.bzl", "ext", dev_dependency = True, isolate = True)
isolated_prox3.use(name = "foo")
prox5 = use_extension("@//bzl:extensions.bzl", "ext")
prox5.use(name = "foo")
prox6 = use_extension("@//bzl:extensions.bzl", "ext", dev_dependency = True)
prox6.use(list = ["foo", "bar"])
prox7 = use_extension("//bzl:extensions.bzl", "ext", dev_dependency = False)
prox7.use(label = "@foo//:bar")
isolated_prox4 = use_extension("@name//bzl:extensions.bzl", "ext", isolate = True)
isolated_prox4.use(name = "foo")
prox8 = use_extension("//bzl:extensions.bzl", "ext", dev_dependency = True)
unused = prox8.use(dict = {"foo": "bar"})
prox9 = use_extension(
    # comment
    "@dep//:extensions.bzl", "other_ext")
prox9.use(label = "@name//:bar")
prox10 = use_extension("@dep//:extensions.bzl", "other_ext", dev_dependency = bool(1))
prox10.use(dict = {"foo": "bar"})
prox11 = use_extension("extension.bzl", "ext")
prox12 = use_extension(":extension.bzl", "ext")
prox13 = use_extension("//:extension.bzl", "ext")
prox14 = use_extension("@name//:extension.bzl", "ext")
prox15 = use_extension("@repo_name//:extension.bzl", "ext")
`

func TestProxies(t *testing.T) {
	for i, tc := range []struct {
		content         string
		extBzlFiles     []string
		extName         string
		dev             bool
		expectedProxies []string
	}{
		{
			proxiesNoModuleHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl"},
			"ext",
			false,
			[]string{"prox5", "prox7"},
		},
		{
			proxiesNoModuleHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl"},
			"ext",
			true,
			[]string{"prox6", "prox8"},
		},
		{
			proxiesModuleNameHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl", "@name//bzl:extensions.bzl"},
			"ext",
			false,
			[]string{"prox1", "prox5", "prox7"},
		},
		{
			proxiesModuleNameHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl", "@name//bzl:extensions.bzl"},
			"ext",
			true,
			[]string{"prox2", "prox6", "prox8"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl", "@repo_name//bzl:extensions.bzl"},
			"ext",
			false,
			[]string{"prox3", "prox5", "prox7"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"//bzl:extensions.bzl", "@//bzl:extensions.bzl", "@repo_name//bzl:extensions.bzl"},
			"ext",
			true,
			[]string{"prox4", "prox6", "prox8"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"@name//bzl:extensions.bzl"},
			"ext",
			false,
			[]string{"prox1"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"@dep//:extensions.bzl"},
			"other_ext",
			false,
			[]string{"prox9"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"@dep//:extensions.bzl"},
			"other_ext",
			true,
			[]string{"prox10"},
		},
		{
			proxiesModuleNameHeader + proxiesBody,
			[]string{"//:extension.bzl", "@//:extension.bzl"},
			"ext",
			false,
			[]string{"prox11", "prox12", "prox13", "prox14"},
		},
		{
			proxiesModuleRepoNameHeader + proxiesBody,
			[]string{"//:extension.bzl", "@//:extension.bzl"},
			"ext",
			false,
			[]string{"prox11", "prox12", "prox13", "prox15"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for _, extBzlFile := range tc.extBzlFiles {
				t.Run("label_"+extBzlFile, func(t *testing.T) {
					f, err := build.ParseModule("MODULE.bazel", []byte(tc.content))
					if err != nil {
						t.Fatal(err)
					}
					actualProxies := Proxies(f, extBzlFile, tc.extName, tc.dev)
					if !reflect.DeepEqual(actualProxies, tc.expectedProxies) {
						t.Error("want: ", tc.expectedProxies, ", got: ", actualProxies)
					}
				})
			}
		})
	}
}

func TestAllProxies(t *testing.T) {
	for i, tc := range []struct {
		proxy           string
		expectedProxies []string
	}{
		{
			"invalid_proxy",
			nil,
		},
		{
			"isolated_prox1",
			[]string{"isolated_prox1"},
		},
		{
			"isolated_prox2",
			[]string{"isolated_prox2"},
		},
		{
			"isolated_prox3",
			[]string{"isolated_prox3"},
		},
		{
			"isolated_prox4",
			[]string{"isolated_prox4"},
		},
		{
			"prox1",
			[]string{"prox1", "prox5", "prox7"},
		},
		{
			"prox2",
			[]string{"prox2", "prox6", "prox8"},
		},
		{
			"prox9",
			[]string{"prox9"},
		},
		{
			"prox10",
			[]string{"prox10"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := build.ParseModule("MODULE.bazel", []byte(proxiesModuleNameHeader+proxiesBody))
			if err != nil {
				t.Fatal(err)
			}
			proxies := AllProxies(f, tc.proxy)
			if !reflect.DeepEqual(proxies, tc.expectedProxies) {
				t.Error("want: ", tc.expectedProxies, ", got: ", proxies)
			}
		})
	}
}

const useReposFile = `
prox1 = use_extension("@mod//bzl:extensions.bzl", "ext")
prox1.use()
prox1.config()
prox2 = use_extension("//bzl:extensions.bzl", "ext")
use_repo(prox2)
use_repo(prox1, "repo5")
prox3 = use_extension("@dep//bzl:extensions.bzl", "ext")
prox2_dev = use_extension("//bzl:extensions.bzl", "ext", dev_dependency = True)
use_repo(prox1, "repo1")
use_repo(prox3, "repo2")
use_repo(prox2, "repo3", "repo4")
`

func TestUseRepos(t *testing.T) {
	for i, tc := range []struct {
		content       string
		proxies       []string
		expectedStmts []int
	}{
		{
			useReposFile,
			[]string{"prox2"},
			[]int{4, 10},
		},
		{
			useReposFile,
			[]string{"prox1", "prox3"},
			[]int{5, 8, 9},
		},
		{
			useReposFile,
			[]string{"prox2_dev"},
			[]int{},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := build.ParseModule("MODULE.bazel", []byte(tc.content))
			if err != nil {
				t.Fatal(err)
			}
			var expectedUseRepos []*build.CallExpr
			for _, stmt := range tc.expectedStmts {
				expectedUseRepos = append(expectedUseRepos, f.Stmt[stmt].(*build.CallExpr))
			}
			actualUseRepos := UseRepos(f, tc.proxies)
			if !reflect.DeepEqual(actualUseRepos, expectedUseRepos) {
				t.Error("want: ", expectedUseRepos, ", got: ", actualUseRepos)
			}
		})
	}
}

func TestNewUseRepos(t *testing.T) {
	for i, tc := range []struct {
		content         string
		proxies         []string
		expectedContent string
		expectedUseRepo int
	}{
		{
			``,
			[]string{"prox1"},
			``,
			-1,
		},
		{
			`prox1 = use_extension("@module//:lib.bzl", "ext")`,
			[]string{"prox2"},
			`prox1 = use_extension("@module//:lib.bzl", "ext")
`,
			-1,
		},
		{
			`prox1 = use_extension("@mod//bzl:extensions.bzl", "ext", dev_dependency = True)`,
			[]string{"prox1"},
			`prox1 = use_extension("@mod//bzl:extensions.bzl", "ext", dev_dependency = True)
use_repo(prox1)
`,
			1,
		},
		{
			`prox1 = use_extension("@mod//bzl:extensions.bzl", "ext")
prox1.config()
prox1.download(name = "foo")`,
			[]string{"prox1"},
			`prox1 = use_extension("@mod//bzl:extensions.bzl", "ext")
prox1.config()
prox1.download(name = "foo")
use_repo(prox1)
`,
			3,
		},
		{
			`go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")

pull = use_extension("@rules_oci//oci:pull.bzl", "go_deps")
pull.oci_pull(name = "distroless_base")
`,
			[]string{"go_deps"},
			`go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
use_repo(go_deps)

pull = use_extension("@rules_oci//oci:pull.bzl", "go_deps")
pull.oci_pull(name = "distroless_base")
`,
			2,
		},
		{
			`go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")

pull = use_extension("@rules_oci//oci:pull.bzl", "go_deps")
`,
			[]string{"go_deps"},
			`go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
use_repo(go_deps)

pull = use_extension("@rules_oci//oci:pull.bzl", "go_deps")
`,
			2,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := build.ParseModule("MODULE.bazel", []byte(tc.content))
			if err != nil {
				t.Fatal(err)
			}
			f, actualNewUseRepo := NewUseRepo(f, tc.proxies)
			actualContent := string(build.Format(f))
			if actualNewUseRepo != nil {
				if !reflect.DeepEqual(actualNewUseRepo, f.Stmt[tc.expectedUseRepo]) {
					t.Error("want: ", f.Stmt[tc.expectedUseRepo], ", got: ", actualNewUseRepo)
				}
			} else {
				if tc.expectedUseRepo != -1 {
					t.Error("wanted a nil new use_repo")
				}
			}
			if !reflect.DeepEqual(actualContent, tc.expectedContent) {
				t.Errorf("want:\n%q\ngot:\n%q\n", tc.expectedContent, actualContent)
			}
		})
	}
}

func TestAddRepoUsages(t *testing.T) {
	for i, tc := range []struct {
		content         string
		repos           []string
		expectedContent string
	}{
		{
			``,
			[]string{},
			``,
		},
		{
			`use_repo(prox)`,
			[]string{"repo2", "repo1"},
			`use_repo(prox, "repo1", "repo2")
`,
		},
		{
			`use_repo(
    prox,
)`,
			[]string{"repo2", "repo1"},
			`use_repo(
    prox,
    "repo1",
    "repo2",
)
`,
		},
		{
			`use_repo(prox, "repo2")`,
			[]string{"repo2", "repo1"},
			`use_repo(prox, "repo1", "repo2")
`,
		},
		{
			`use_repo(
    prox,
    "repo2",
)`,
			[]string{"repo2", "repo1"},
			`use_repo(
    prox,
    "repo1",
    "repo2",
)
`,
		},
		{
			`use_repo(prox, "repo1")
use_repo(
    prox2,
    my_repo = "repo2",
    "repo5",
)
use_repo(prox, "repo3")`,
			[]string{"repo3", "repo2", "repo4"},
			`use_repo(prox, "repo1")

use_repo(
    prox2,
    "repo5",
    my_repo = "repo2",
)

use_repo(prox, "repo3", "repo4")
`,
		},
		// Test mapped repo names with valid identifiers
		{
			`use_repo(prox)`,
			[]string{"foo", "bar=baz"},
			`use_repo(prox, "foo", bar = "baz")
`,
		},
		{
			`use_repo(prox)`,
			[]string{"my_repo=actual_repo"},
			`use_repo(prox, my_repo = "actual_repo")
`,
		},
		// Test mapped repo names with invalid identifiers (dict unpacking)
		{
			`use_repo(prox)`,
			[]string{"foo.2=foo"},
			`use_repo(prox, **{"foo.2": "foo"})
`,
		},
		{
			`use_repo(prox)`,
			[]string{"foo-bar=baz"},
			`use_repo(prox, **{"foo-bar": "baz"})
`,
		},
		// Test mixed positional, keyword, and dict unpacking
		{
			`use_repo(prox)`,
			[]string{"simple", "valid_key=value", "invalid.key=value2"},
			`use_repo(prox, "simple", valid_key = "value", **{"invalid.key": "value2"})
`,
		},
		// Test extending existing dict unpacking
		{
			`use_repo(prox, **{"existing.key": "existing_value"})`,
			[]string{"new.key=new_value"},
			`use_repo(prox, **{
    "existing.key": "existing_value",
    "new.key": "new_value",
})
`,
		},
		// Test that duplicates are not added
		{
			`use_repo(prox, "repo1")`,
			[]string{"repo1", "repo2"},
			`use_repo(prox, "repo1", "repo2")
`,
		},
		{
			`use_repo(prox, my_repo = "actual")`,
			[]string{"my_other=actual", "new=other"},
			`use_repo(prox, my_other = "actual", new = "other")
`,
		},
		// Test that existing mappings are preserved when adding the same underlying repo
		{
			`use_repo(image, my_ubuntu = "img_12345")`,
			[]string{"img_12345"},
			`use_repo(image, my_ubuntu = "img_12345")
`,
		},
		{
			`use_repo(ext, custom_name = "repo_value")`,
			[]string{"repo_value", "other_repo"},
			`use_repo(ext, "other_repo", custom_name = "repo_value")
`,
		},
		{
			`use_repo(ext, my_mapping = "actual_repo", "existing")`,
			[]string{"actual_repo", "new_repo"},
			`use_repo(ext, "existing", "new_repo", my_mapping = "actual_repo")
`,
		},
		// Test that mappings can be replaced/renamed
		{
			`use_repo(proxy, foo = "bar")`,
			[]string{"foo=baz"},
			`use_repo(proxy, foo = "baz")
`,
		},
		{
			`use_repo(proxy, foo = "bar")`,
			[]string{"qux=bar"},
			`use_repo(proxy, qux = "bar")
`,
		},
		{
			`use_repo(proxy, foo = "bar", "other")`,
			[]string{"foo=baz"},
			`use_repo(proxy, "other", foo = "baz")
`,
		},
		{
			`use_repo(proxy, foo = "bar", "other")`,
			[]string{"qux=bar", "new"},
			`use_repo(proxy, "other", "new", qux = "bar")
`,
		},
		// Test that Starlark reserved keywords are treated as invalid identifiers
		{
			`use_repo(proxy)`,
			[]string{"for=my_repo"},
			`use_repo(proxy, **{"for": "my_repo"})
`,
		},
		{
			`use_repo(proxy)`,
			[]string{"if=my_repo", "else=other"},
			`use_repo(proxy, **{
    "if": "my_repo",
    "else": "other",
})
`,
		},
		// Test that True, False, None are also treated as reserved keywords
		{
			`use_repo(proxy)`,
			[]string{"True=my_repo"},
			`use_repo(proxy, **{"True": "my_repo"})
`,
		},
		{
			`use_repo(proxy)`,
			[]string{"False=repo1", "None=repo2"},
			`use_repo(proxy, **{
    "False": "repo1",
    "None": "repo2",
})
`,
		},
		// Test that foo=foo is simplified to just foo (positional argument)
		{
			`use_repo(proxy)`,
			[]string{"foo=foo"},
			`use_repo(proxy, "foo")
`,
		},
		{
			`use_repo(proxy)`,
			[]string{"foo=foo", "bar=baz"},
			`use_repo(proxy, "foo", bar = "baz")
`,
		},
		{
			`use_repo(proxy)`,
			[]string{"foo=foo", "bar=bar", "baz=qux"},
			`use_repo(proxy, "bar", "foo", baz = "qux")
`,
		},
		// Test foo=foo with existing repos
		{
			`use_repo(proxy, "existing")`,
			[]string{"foo=foo"},
			`use_repo(proxy, "existing", "foo")
`,
		},
		// Test that non-ASCII identifiers use dict unpacking (ASCII-only validation)
		{
			`use_repo(proxy)`,
			[]string{"café=my_repo"},
			`use_repo(proxy, **{"caf\303\251": "my_repo"})
`,
		},
		{
			`use_repo(proxy)`,
			[]string{"变量=var"},
			`use_repo(proxy, **{"\345\217\230\351\207\217": "var"})
`,
		},
		// Test support for Bazel placeholders in repository names.
		// See https://github.com/bazelbuild/bazel/pull/27890 for the Bazel feature.
		// Placeholders like {name} and {version} in the RHS are replaced with module attributes by Bazel.
		// buildozer should preserve these placeholders when adding repos via use_repo_add.
		{
			`use_repo(ext)`,
			[]string{"custom={name}_suffix"},
			`use_repo(ext, custom = "{name}_suffix")
`,
		},
		{
			`use_repo(ext)`,
			[]string{"my_repo={name}-v{version}"},
			`use_repo(ext, my_repo = "{name}-v{version}")
`,
		},
		{
			`use_repo(ext)`,
			[]string{"local_name={version}_tag"},
			`use_repo(ext, local_name = "{version}_tag")
`,
		},
		// Test that placeholders already present in use_repo are preserved (not duplicated)
		{
			`use_repo(ext, custom = "{name}-v{version}")`,
			[]string{"custom={name}-v{version}"},
			`use_repo(ext, custom = "{name}-v{version}")
`,
		},
		{
			`use_repo(ext, my_mapping = "{name}_value")`,
			[]string{"my_mapping={name}_value"},
			`use_repo(ext, my_mapping = "{name}_value")
`,
		},
		// Test mixed scenarios with placeholders and regular repos
		{
			`use_repo(ext)`,
			[]string{"regular_repo", "alias={version}_tag"},
			`use_repo(ext, "regular_repo", alias = "{version}_tag")
`,
		},
		{
			`use_repo(ext, "existing")`,
			[]string{"custom={name}-{version}", "other"},
			`use_repo(ext, "existing", "other", custom = "{name}-{version}")
`,
		},
		{
			`use_repo(ext, "foo")`,
			[]string{"versioned={name}-{version}", "bar"},
			`use_repo(ext, "foo", "bar", versioned = "{name}-{version}")
`,
		},
		// Test that existing placeholder mappings can be updated/replaced
		{
			`use_repo(ext, my_name = "{name}_old")`,
			[]string{"my_name={name}_new"},
			`use_repo(ext, my_name = "{name}_new")
`,
		},
		{
			`use_repo(ext, old_key = "{name}-v{version}")`,
			[]string{"new_key={name}-v{version}"},
			`use_repo(ext, new_key = "{name}-v{version}")
`,
		},
		// Test placeholders survive when the same value already exists under different mapping
		{
			`use_repo(ext, version1 = "{name}-v{version}")`,
			[]string{"{name}-v{version}"},
			`use_repo(ext, version1 = "{name}-v{version}")
`,
		},
		{
			`use_repo(ext, old_mapping = "{name}_suffix")`,
			[]string{"new_mapping={name}_suffix"},
			`use_repo(ext, new_mapping = "{name}_suffix")
`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := build.ParseModule("MODULE.bazel", []byte(tc.content))
			if err != nil {
				t.Fatal(err)
			}
			var useRepos []*build.CallExpr
			for _, stmt := range f.Stmt {
				useRepos = append(useRepos, stmt.(*build.CallExpr))
			}
			AddRepoUsages(useRepos, tc.repos...)
			actualContent := string(build.Format(f))
			if !reflect.DeepEqual(actualContent, tc.expectedContent) {
				t.Errorf("want:\n%q\ngot:\n%q\n", tc.expectedContent, actualContent)
			}
		})
	}
}

func TestRemoveRepoUsages(t *testing.T) {
	for i, tc := range []struct {
		content         string
		repos           []string
		expectedContent string
	}{
		{
			``,
			[]string{"repo1"},
			``,
		},
		{
			`use_repo(prox)`,
			[]string{"repo2", "repo1"},
			`use_repo(prox)
`,
		},
		{
			`use_repo(
    prox,
)`,
			[]string{"repo2", "repo1"},
			`use_repo(
    prox,
)
`,
		},
		{
			`use_repo(prox, "repo2")`,
			[]string{"repo2", "repo1"},
			`use_repo(prox)
`,
		},
		{
			`use_repo(
    prox,
    "repo2",
)`,
			[]string{"repo2", "repo1"},
			`use_repo(prox)
`,
		},
		{
			`use_repo(prox, "repo1")
use_repo(
    prox2,
    my_repo = "repo2",
    "repo5",
)
use_repo(prox, "repo3")`,
			[]string{"repo3", "repo2", "repo4"},
			`use_repo(prox, "repo1")

use_repo(
    prox2,
    "repo5",
)

use_repo(prox)
`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := build.ParseModule("MODULE.bazel", []byte(tc.content))
			if err != nil {
				t.Fatal(err)
			}
			var useRepos []*build.CallExpr
			for _, stmt := range f.Stmt {
				useRepos = append(useRepos, stmt.(*build.CallExpr))
			}
			RemoveRepoUsages(useRepos, tc.repos...)
			actualContent := string(build.Format(f))
			if !reflect.DeepEqual(actualContent, tc.expectedContent) {
				t.Errorf("want:\n%q\ngot:\n%q\n", tc.expectedContent, actualContent)
			}
		})
	}
}
