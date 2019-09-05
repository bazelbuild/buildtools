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
			`:1: Module "//foo/bar/internal/baz:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package".`,
			`:2: Module "//foo/bar/private/baz:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package".`,
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
			`:1: Module "//foo/bar/internal:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package".`,
			`:2: Module "//foo/bar/private:module.bzl" can only be loaded from files located inside "//foo/bar", not from "//test/package".`,
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
}