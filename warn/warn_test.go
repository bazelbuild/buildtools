package warn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/testutils"
)

const (
	scopeBuild      = build.TypeBuild
	scopeBzl        = build.TypeDefault
	scopeWorkspace  = build.TypeWorkspace
	scopeEverywhere = scopeBuild | scopeBzl | scopeWorkspace
)

func getFilename(fileType build.FileType) string {
	switch fileType {
	case build.TypeBuild:
		return "BUILD"
	case build.TypeWorkspace:
		return "WORKSPACE"
	default:
		return "test_file.bzl"
	}
}

func getFindings(category, input string, fileType build.FileType) []*Finding {
	input = strings.TrimLeft(input, "\n")
	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return FileWarnings(buildFile, "the_package", []string{category}, false)
}

func compareFindings(t *testing.T, category, input string, expected []string, scope, fileType build.FileType) {
	// If scope doesn't match the file type, no warnings are expected
	if scope&fileType == 0 {
		expected = []string{}
	}

	findings := getFindings(category, input, fileType)
	// We ensure that there is the expected number of warnings.
	// At the moment, we check only the line numbers.
	if len(expected) != len(findings) {
		t.Errorf("Input: %s", input)
		t.Errorf("number of matches: %d, want %d", len(findings), len(expected))
		for _, e := range expected {
			t.Errorf("expected: %s", e)
		}
		for _, f := range findings {
			t.Errorf("got: %d: %s", f.Start.Line, f.Message)
		}
		return
	}
	for i := range findings {
		msg := fmt.Sprintf(":%d: %s", findings[i].Start.Line, findings[i].Message)
		if !strings.Contains(msg, expected[i]) {
			t.Errorf("Input: %s", input)
			t.Errorf("got:  `%s`,\nwant: `%s`", msg, expected[i])
		}
	}
}

func checkFix(t *testing.T, category, input, expected string, scope, fileType build.FileType) {
	// If scope doesn't match the file type, no changes are expected
	if scope&fileType == 0 {
		expected = input
	}

	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	goldenFile, err := build.Parse(getFilename(fileType), []byte(expected))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	FixWarnings(buildFile, "the_package", []string{category}, false)
	have := build.Format(buildFile)
	want := build.Format(goldenFile)
	if !bytes.Equal(have, want) {
		t.Errorf("fixed a test (type %s) incorrectly:\ninput:\n%s\ndiff (-expected, +ours)\n",
			fileType, input)
		testutils.Tdiff(t, want, have)
	}
}

func checkFindings(t *testing.T, category, input string, expected []string, scope build.FileType) {
	// The same as checkFindingsAndFix but ensure that fixes don't change the file (except for formatting)
	checkFindingsAndFix(t, category, input, input, expected, scope)
}

func checkFindingsAndFix(t *testing.T, category, input, output string, expected []string, scope build.FileType) {
	fileTypes := []build.FileType{
		build.TypeDefault,
		build.TypeBuild,
		build.TypeWorkspace,
	}

	for _, fileType := range fileTypes {
		compareFindings(t, category, input, expected, scope, fileType)
		checkFix(t, category, input, output, scope, fileType)
		checkFix(t, category, output, output, scope, fileType)
	}
}
