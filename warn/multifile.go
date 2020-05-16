package warn

import (
	"github.com/bazelbuild/buildtools/build"
	"strings"
)

// FileReader is a class that can read an arbitrary Starlark file
// from the repository and cache the results.
type FileReader struct {
	cache    map[string]*build.File
	readFile func(string) ([]byte, error)
}

// NewFileReader creates and initializes a FileReader instance with a
// custom readFile function that can read an arbitrary file in the
// repository using a path relative to the workspace root
// (OS-independent, with forward slashes).
func NewFileReader(readFile func(string) ([]byte, error)) *FileReader {
	return &FileReader{
		cache:    make(map[string]*build.File),
		readFile: readFile,
	}
}

// retrieveFile reads a Starlark file using only the readFile method
// (without using the cache).
func (fr *FileReader) retrieveFile(filename string) *build.File {
	contents, err := fr.readFile(filename)
	if err != nil {
		return nil
	}

	file, err := build.Parse(filename, contents)
	if err != nil {
		return nil
	}
	file.Path = filename

	return file
}

// GetFile reads a Starlark file from the repository or the cache.
// Returns nil if the file is not found or not valid.
func (fr *FileReader) GetFile(pkg, label string) *build.File {
	filename := label
	if pkg != "" {
		filename = pkg + "/" + label
	}

	// Try to retrieve from the cache
	if file, ok := fr.cache[filename]; ok {
		return file
	}
	file := fr.retrieveFile(filename)
	if file != nil {
		file.Pkg = pkg
		file.Label = label
	}
	fr.cache[filename] = file
	return file
}

// ResolveLabel resolves a label to a .bzl file (which may be absolute or relative to the current package)
// to a pair (package, label) for the .bzl file
func ResolveLabel(currentPkg, label string) (pkg, newLabel string) {
	switch {
	case strings.HasPrefix(label, "//"):
		// Absolute label path
		label = label[2:]
		if chunks := strings.SplitN(label, ":", 2); len(chunks) == 2 {
			return chunks[0], chunks[1]
		}
		return "", label
	case strings.HasPrefix(label, ":"):
		// Relative label path
		return currentPkg, label[1:]
	default:
		// External repositories are not supported
		return "", ""
	}
}
