package warn

import (
	"github.com/bazelbuild/buildtools/build"
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
		cache: make(map[string]*build.File),
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

	return file
}

// GetFile reads a Starlark file from the repository or the cache.
// Returns nil if the file is not found or not valid.
func (fr *FileReader) GetFile(filename string) *build.File {
	// Try to retrieve from the cache
	if file, ok := fr.cache[filename]; ok {
		return file
	}
	file := fr.retrieveFile(filename)
	fr.cache[filename] = file
	return file
}
