package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

// FindTests finds all files matching the given pattern.
// It changes the working directory to `directory`,  and returns a function
// to call to change back to the current directory.
// This allows tests to assert on alias finding between absolute and relative labels.
func FindTests(t *testing.T, directory, pattern string) ([]string, func()) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), directory)); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("Didn't find any test cases")
	}
	return files, func() { os.Chdir(wd) }
}
