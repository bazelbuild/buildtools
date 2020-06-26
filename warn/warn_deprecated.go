// Warnings for using deprecated functions

package warn

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"
)

func checkDeprecatedFunction(stmt build.Expr, loadedSymbols *map[string]*build.Ident, fullLabel string) *LinterFinding {
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
	str, ok := (*docstring).(*build.StringExpr)
	if !ok {
		return nil
	}
	docstringInfo := parseFunctionDocstring(str)
	if !docstringInfo.deprecated {
		return nil
	}

	return makeLinterFinding(node, fmt.Sprintf("The function %q defined in %q is deprecated.", def.Name, fullLabel))
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
		pkg, fileLabel := ResolveLabel(f.Pkg, load.Module.Value)
		if fileLabel == "" {
			continue
		}
		loadedFile := fileReader.GetFile(pkg, fileLabel)
		if loadedFile == nil {
			continue
		}
		loadedSymbols := make(map[string]*build.Ident)
		for _, from := range load.From {
			loadedSymbols[from.Name] = from
		}

		for _, stmt := range loadedFile.Stmt {
			if finding := checkDeprecatedFunction(stmt, &loadedSymbols, loadedFile.CanonicalPath()); finding != nil {
				findings = append(findings, finding)
			}
		}

	}
	return findings
}
