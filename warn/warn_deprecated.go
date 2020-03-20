// Warnings for using deprecated functions

package warn

import (
	"fmt"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

func getPathFromLabel(label, pkg string) string {
	switch {
	case strings.HasPrefix(label, "//"):
		// Absolute label path
		return strings.ReplaceAll(label[2:], ":", "/")
	case strings.HasPrefix(label, ":"):
		// Relative label path
		return pkg + "/" + label[1:]
	default:
		// External repositories are not supported
		return ""
	}
}

func readBzlFile(path string, fileReader *FileReader) (*build.File, bool) {
	file := fileReader.GetFile(path)
	return file, file != nil
}

func checkDeprecatedFunction(stmt build.Expr, loadedSymbols *map[string]*build.Ident, path string) *LinterFinding {
	def, ok := stmt.(*build.DefStmt)
	if !ok {
		return nil
	}
	node, ok := (*loadedSymbols)[def.Name]
	if !ok {
		return nil
	}
	docstring, ok := getDocstring(def.Body)
	if !ok {
		return nil
	}
	str, ok := docstring.(*build.StringExpr)
	if !ok {
		return nil
	}
	docstringInfo := parseFunctionDocstring(str)
	if !docstringInfo.deprecated {
		return nil
	}

	return makeLinterFinding(node, fmt.Sprintf("The function %q defined in %q is deprecated.", def.Name, path))
}

func deprecatedFunctionWarning(f *build.File, fileReader *FileReader) []*LinterFinding {
	if fileReader == nil {
		return nil
	}

	findings := []*LinterFinding{}
	for _, stmt := range f.Stmt {
		load, ok := stmt.(*build.LoadStmt)
		if !ok {
			continue
		}
		path := getPathFromLabel(load.Module.Value, f.Pkg)
		if path == "" {
			continue
		}
		loadedFile, ok := readBzlFile(path, fileReader)
		if !ok {
			continue
		}

		loadedSymbols := make(map[string]*build.Ident)
		for _, from := range load.From {
			loadedSymbols[from.Name] = from
		}

		for _, stmt := range loadedFile.Stmt {
			if finding := checkDeprecatedFunction(stmt, &loadedSymbols, path); finding != nil {
				findings = append(findings, finding)
			}
		}

	}
	return findings
}
