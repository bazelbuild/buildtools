package main

import (
	"os"
	"path/filepath"
	"strings"
)

func isStarlarkFile(filename string) bool {
	basename := strings.ToLower(filepath.Base(filename))
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

// ExpandDirectories takes a list of file/directory names and returns a list with file names
// by traversing each directory recursively and searching for relevant Starlark files.
func ExpandDirectories(args []string) ([]string, error) {
	files := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return []string{}, err
		}
		if !info.IsDir() {
			files = append(files, arg)
			continue
		}
		err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if isStarlarkFile(path) {
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
