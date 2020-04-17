// Package utils contains shared methods that can be used by different implementations of
// buildifier binary
package utils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/warn"
)

func isStarlarkFile(name string) bool {
	ext := filepath.Ext(name)
	switch ext {
	case ".bzl", ".sky":
		return true
	}

	switch ext {
	case ".bazel", ".oss":
		// BUILD.bazel or BUILD.foo.bazel should be treated as Starlark files, same for WORSKSPACE
		return strings.HasPrefix(name, "BUILD.") || strings.HasPrefix(name, "WORKSPACE.")
	}

	return name == "BUILD" || name == "WORKSPACE"
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
			if !info.IsDir() && isStarlarkFile(info.Name()) {
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

// containsFile checks whether a given directory contains a file called
// <file> or <file>.bazel.
func containsFile(dir, file string) bool {
	for _, filename := range []string{file, file + ".bazel"} {
		path := filepath.Join(dir, filename)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

// SplitRelativePath splits a path relative to the workspace root into package name and label.
// It takes the workspace root and chunks of the relative path (already split) as input arguments.
// Both output variables always have forward slashes as separators.
func SplitRelativePath(workspaceRoot string, chunks []string) (pkg, label string) {
	switch chunks[len(chunks)-1] {
	case "BUILD", "BUILD.bazel":
		return path.Join(chunks[:len(chunks)-1]...), chunks[len(chunks)-1]
	}

	pkg = ""
	label = path.Join(chunks...)
	parent := workspaceRoot
	for i, chunk := range chunks {
		if i == len(chunks)-1 {
			// The last chunk is a filename, not a directory
			break
		}
		parent = filepath.Join(parent, chunk)
		if containsFile(parent, "BUILD") {
			pkg = path.Join(chunks[:i+1]...)
			label = path.Join(chunks[i+1:]...)
		}
	}
	return pkg, label
}

// SplitFilePath splits a file path into the workspace root, package name and label.
// Workspace root is determined as the last directory in the file path that
// contains a WORKSPACE (or WORKSPACE.bazel) file.
// Package and label are always separated with forward slashes.
// Returns empty strings if no WORKSPACE file is found.
func SplitFilePath(filename string) (workspaceRoot, pkg, label string) {
	root := "/"
	if volume := filepath.VolumeName(filename); volume != "" {
		// Windows
		root = volume + "\\"
	}
	// filename relative to the file system root
	relPath := filename[len(root):]

	chunks := append([]string{""}, strings.Split(relPath, string(os.PathSeparator))...)
	parent := root
	workspaceIndex := -1
	for i, chunk := range chunks {
		if i == len(chunks)-1 {
			// The last chunk is a filename, not a directory
			break
		}
		parent = filepath.Join(parent, chunk)
		if containsFile(parent, "WORKSPACE") {
			workspaceRoot = parent
			workspaceIndex = i
		}
	}
	if workspaceIndex != -1 {
		pkg, label = SplitRelativePath(workspaceRoot, chunks[workspaceIndex+1:])
	}
	return workspaceRoot, pkg, label
}

// getFileReader returns a *FileReader object that reads files from the local
// filesystem if the workspace root is known.
func getFileReader(workspaceRoot string) *warn.FileReader {
	if workspaceRoot == "" {
		return nil
	}

	readFile := func(filename string) ([]byte, error) {
		// Use OS-specific path separators
		filename = strings.ReplaceAll(filename, "/", string(os.PathSeparator))
		path := filepath.Join(workspaceRoot, filename)

		return ioutil.ReadFile(path)
	}

	return warn.NewFileReader(readFile)
}

// Lint calls the linter and returns a list of unresolved findings
func Lint(f *build.File, lint string, warningsList *[]string, verbose bool) []*warn.Finding {
	fileReader := getFileReader(f.WorkspaceRoot)

	switch lint {
	case "warn":
		return warn.FileWarnings(f, *warningsList, nil, warn.ModeWarn, fileReader)
	case "fix":
		warn.FixWarnings(f, *warningsList, verbose, fileReader)
	}
	return nil
}
