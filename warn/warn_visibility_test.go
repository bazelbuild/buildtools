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
