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

package warn

import "testing"

func TestBzlVisibility(t *testing.T) {
	checkFindings(t, "bzl-visibility", `
load("//foo/bar/internal/baz:module.bzl", "foo")
load("//foo/bar/private/baz:module.bzl", "bar")
load("//foo/bar/external/baz:module.bzl", "baz")

foo()
bar()
baz()
`,
		[]string{
			`:1: Module "//foo/bar/internal/baz:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
			`:2: Module "//foo/bar/private/baz:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
		},
		scopeEverywhere)

	checkFindings(t, "bzl-visibility", `
load("//foo/bar/internal:module.bzl", "foo")
load("//foo/bar/private:module.bzl", "bar")
load("//foo/bar/external:module.bzl", "baz")

foo()
bar()
baz()
`,
		[]string{
			`:1: Module "//foo/bar/internal:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
			`:2: Module "//foo/bar/private:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
		},
		scopeEverywhere)

	checkFindings(t, "bzl-visibility", `
load("@repo//foo/bar/internal:module.bzl", "foo")
load("@repo//foo/bar/private:module.bzl", "bar")
load("@repo//foo/bar/external:module.bzl", "baz")
load("@repo/internal:module.bzl", "qux")  # incorrect, but shouldn't cause buildifier crashes


foo()
bar()
baz()
qux()
`,
		[]string{
			`:1: Module "@repo//foo/bar/internal:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
			`:2: Module "@repo//foo/bar/private:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package/`,
			`:4: Module "@repo/internal:module.bzl" can only be loaded from files located inside "@repo", not from "//test/package/`,
		},
		scopeEverywhere)

	checkFindings(t, "bzl-visibility", `
load("//test/internal/foo:module.bzl", "foo")
load("//test/private/foo:module.bzl", "bar")
load("//test/external/foo:module.bzl", "baz")
load("//test/internal:module.bzl", "foo")
load("//test/private:module.bzl", "bar")
load("//test/external:module.bzl", "baz")
`,
		[]string{},
		scopeEverywhere)

	checkFindings(t, "bzl-visibility", `
load("@repo//test/internal/foo:module.bzl", "foo")
load("@repo//test/private/foo:module.bzl", "bar")
load("@repo//test/external/foo:module.bzl", "baz")
load("@repo//test/internal:module.bzl", "foo")
load("@repo//test/private:module.bzl", "bar")
load("@repo//test/external:module.bzl", "baz")
`,
		[]string{},
		scopeEverywhere)
}

func TestBzlVisibilityJavatest(t *testing.T) {
	defer setUpTestPackage("foo/javatests/bar")()

	checkFindings(t, "bzl-visibility", `
load("//foo/java/bar/internal/baz:module.bzl", "foo")
load("//foo/java/bar/private/baz:module.bzl", "bar")
load("//foo/javatests/bar/internal/baz:module.bzl", "foo1")
load("//foo/javatests/bar/private/baz:module.bzl", "bar1")

foo()
bar()
foo1()
bar1()
`,
		[]string{},
		scopeEverywhere)
}
