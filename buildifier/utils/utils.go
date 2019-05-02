// Package utils contains shared methods that can be used by different implementations of
// buildifier binary
package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/warn"
)

func isStarlarkFile(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	basename := strings.ToLower(info.Name())
	ext := filepath.Ext(basename)
	switch ext {
	case ".bzl", ".sky":
		return true
	}
	base := basename[:len(basename)-len(ext)]
	switch {
	case ext == ".build" || base == "build":
		return true
	case ext == ".workspace" || base == "workspace":
		return true
	}
	return false
}

func skip(info os.FileInfo) bool {
	return info.IsDir() && info.Name() == ".git"
}

// ExpandDirectories takes a list of file/directory names and returns a list with file names
// by traversing each directory recursively and searching for relevant Starlark files.
func ExpandDirectories(args *[]string) ([]string, error) {
	files := []string{}
	for _, arg := range *args {
		info, err := os.Stat(arg)
		if err != nil {
			return []string{}, err
		}
		if !info.IsDir() {
			files = append(files, arg)
			continue
		}
		err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if skip(info) {
				return filepath.SkipDir
			}
			if isStarlarkFile(info) {
				files = append(files, path)
			}
			return err
		})
		if err != nil {
			return []string{}, err
		}
	}
	return files, nil
}

// GetParser returns a parser for a given file type
func GetParser(inputType string) func(filename string, data []byte) (*build.File, error) {
	switch inputType {
	case "build":
		return build.ParseBuild
	case "bzl":
		return build.ParseBzl
	case "auto":
		return build.Parse
	case "workspace":
		return build.ParseWorkspace
	default:
		return build.ParseDefault
	}
}

// GetPackageName returns the package name of a file by searching for a WORKSPACE file
func GetPackageName(filename string) string {
	dirs := filepath.SplitList(path.Dir(filename))
	parent := ""
	index := len(dirs) - 1
	for i, chunk := range dirs {
		parent = path.Join(parent, chunk)
		metadata := path.Join(parent, "METADATA")
		if _, err := os.Stat(metadata); !os.IsNotExist(err) {
			index = i
		}
	}
	return strings.Join(dirs[index+1:], "/")
}

// Lint calls the linter and returns a list of unresolved findings
func Lint(f *build.File, pkg, lint string, warningsList *[]string, verbose bool) []*warn.Finding {
	switch lint {
	case "warn":
		return warn.FileWarnings(f, pkg, *warningsList, false)
	case "fix":
		warn.FixWarnings(f, pkg, *warningsList, verbose)
	}
	return nil
}
