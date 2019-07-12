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
	scopeBzl        = build.TypeBzl
	scopeWorkspace  = build.TypeWorkspace
	scopeDefault    = build.TypeDefault
	scopeEverywhere = scopeBuild | scopeBzl | scopeWorkspace | scopeDefault
	scopeBazel      = scopeBuild | scopeBzl | scopeWorkspace
)

func getFilename(fileType build.FileType) string {
	switch fileType {
	case build.TypeBuild:
		return "package/BUILD"
	case build.TypeWorkspace:
		return "package/WORKSPACE"
	case build.TypeBzl:
		return "package/test_file.bzl"
	default:
		return "test_file.strlrk"
	}
}

func getFindings(category, input string, fileType build.FileType) []*Finding {
	input = strings.TrimLeft(input, "\n")
	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return FileWarnings(buildFile, "the_package", []string{category}, nil, ModeWarn)
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

// checkFix makes sure that fixed file contents match the expected output
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

// checkFix makes sure that the file contents don't change if a fix is not requested
// (i.e. the warning functions have no side effects modifying the AST)
func checkNoFix(t *testing.T, category, input string, fileType build.FileType) {
	buildFile, err := build.Parse(getFilename(fileType), []byte(input))
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	formatted := build.Format(buildFile)

	// No fixes expected
	FileWarnings(buildFile, "the_package", []string{category}, nil, ModeWarn)
	fixed := build.Format(buildFile)

	if !bytes.Equal(formatted, fixed) {
		t.Errorf("Modified a file (type %s) while getting warnings:\ninput:\n%s\ndiff (-before, +after)\n",
			fileType, input)
		testutils.Tdiff(t, formatted, fixed)
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
		build.TypeBzl,
	}

	for _, fileType := range fileTypes {
		compareFindings(t, category, input, expected, scope, fileType)
		checkFix(t, category, input, output, scope, fileType)
		checkFix(t, category, output, output, scope, fileType)
		checkNoFix(t, category, input, fileType)
	}
}

func TestCalculateDifference(t *testing.T) {
	tests := []struct {
		before      string
		after       string
		start       int
		end         int
		replacement string
	}{
		{
			before:      "asdf",
			after:       "asxydf",
			start:       2,
			end:         2,
			replacement: "xy",
		},
		{
			before:      "asxydf",
			after:       "asdf",
			start:       2,
			end:         4,
			replacement: "",
		},
		{
			before:      "asxydf",
			after:       "asztdf",
			start:       2,
			end:         4,
			replacement: "zt",
		},
		{
			before:      "",
			after:       "foobar",
			start:       0,
			end:         0,
			replacement: "foobar",
		},
		{
			before:      "foobar",
			after:       "",
			start:       0,
			end:         6,
			replacement: "",
		},
		{
			before:      "qwerty",
			after:       "asdfgh",
			start:       0,
			end:         6,
			replacement: "asdfgh",
		},
		{
			before:      "aa",
			after:       "aaaa",
			start:       2,
			end:         2,
			replacement: "aa",
		},
		{
			before:      "aaaa",
			after:       "aa",
			start:       2,
			end:         4,
			replacement: "",
		},
		{
			before:      "abc",
			after:       "abdbc",
			start:       2,
			end:         2,
			replacement: "db",
		},
		{
			before:      "abdbc",
			after:       "abc",
			start:       2,
			end:         4,
			replacement: "",
		},
	}

	for _, tc := range tests {
		before := []byte(tc.before)
		after := []byte(tc.after)

		start, end, replacement := calculateDifference(&before, &after)

		if start != tc.start || end != tc.end || replacement != tc.replacement {
			t.Errorf("Wrong difference for %q and %q: want %d, %d, %q, got %d, %d, %q",
				tc.before, tc.after, tc.start, tc.end, tc.replacement, start, end, replacement)
		}
	}
}

func TestSuggestions(t *testing.T) {
	// Suggestions are not generated by individual warning functions but by the warnings framework.

	contents := `foo()

attr.bar(name = "bar", cfg = "data")

attr.baz("baz", cfg = "data")
`
	f, err := build.ParseBzl("file.bzl", []byte(contents))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	findings := FileWarnings(f, "pkg", []string{"attr-cfg"}, nil, ModeSuggest)
	want := []struct {
		start       int
		end         int
		replacement string
	}{
		{
			start:       28,
			end:         42,
			replacement: "",
		},
		{
			start:       59,
			end:         73,
			replacement: "",
		},
	}

	if len(findings) != len(want) {
		t.Errorf("Expected %d findings, got %d", len(want), len(findings))
	}

	for i, f := range findings {
		w := want[i]
		if f.Replacement == nil {
			t.Errorf("No replacement for finding %d", i)
		}
		r := f.Replacement
		if r.Start != w.start || r.End != w.end || r.Content != w.replacement {
			t.Errorf("Wrong replacement #%d, want %d, %d, %q, got %d, %d, %q",
				i, w.start, w.end, w.replacement, r.Start, r.End, r.Content)
		}
	}
}

func TestDisabledWarning(t *testing.T) {
	contents := `foo()

# buildifier: disable=depset-iteration
for x in depset([1, 2, 3]):
    print(x)  # buildozer: disable=print

for y in "foobar":  # buildozer: disable=string-iteration
    # buildifier: disable=no-effect
    y

# buildifier: disable=duplicated-name-2
cc_library(
   name = "foo",  # buildifier: disable=duplicated-name-1
)
`

	f, err := build.ParseBzl("file.bzl", []byte(contents))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tests := []struct {
		start    int
		end      int
		category string
	}{
		{
			start:    4,
			end:      5,
			category: "depset-iteration",
		},
		{
			start:    5,
			end:      5,
			category: "print",
		},
		{
			start:    7,
			end:      7,
			category: "string-iteration",
		},
		{
			start:    9,
			end:      9,
			category: "no-effect",
		},
		{
			start:    13,
			end:      13,
			category: "duplicated-name-1",
		},
		{
			start:    12,
			end:      14,
			category: "duplicated-name-2",
		},
	}

	linesCount := strings.Count(contents, "\n")

	for _, tc := range tests {
		for line := 1; line <= linesCount; line++ {
			disabled := DisabledWarning(f, line, tc.category)
			shouldBeDisabled := line >= tc.start && line <= tc.end
			if disabled != shouldBeDisabled {
				t.Errorf("Wrong disabled status for the category %q, want %t, got %t", tc.category, shouldBeDisabled, disabled)
			}
		}
	}
}
