package edit

import (
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
)

func TestMovePackageDeclarationToTheTop(t *testing.T) {
	tests := []struct {
		input, expected string
		shouldMove      bool
	}{
		{`"""Docstring."""

load(":path.bzl", "x")

foo()

package(attr = "val")`,
			`"""Docstring."""

load(":path.bzl", "x")

package(attr = "val")

foo()`,
			true},
		{`"""Docstring."""

load(":path.bzl", "x")

package(attr = "val")

foo()`,
			`"""Docstring."""

load(":path.bzl", "x")

package(attr = "val")

foo()`,
			false},
		{`"""Docstring."""

load(":path.bzl", "x")

foo()`,
			`"""Docstring."""

load(":path.bzl", "x")

foo()`,
			false,
		},
	}

	for _, tst := range tests {
		bld, err := build.Parse("BUILD", []byte(tst.input))
		if err != nil {
			t.Error(err)
			continue
		}
		if result := movePackageDeclarationToTheTop(bld); result != tst.shouldMove {
			t.Errorf("TestMovePackageDeclarationToTheTop: expected %v, got %v", tst.shouldMove, result)
		}

		got := strings.TrimSpace(string(build.Format(bld)))
		want := strings.TrimSpace(tst.expected)

		if got != want {
			t.Errorf("TestMovePackageDeclarationToTheTop: got:\n%s\nexpected:\n%s", got, want)
		}
	}
}
