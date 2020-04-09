// Warnings for using deprecated functions

package warn

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"
)

// Internal constant that represents the native module
const nativeModule = "<native>"

// acceptsNameArgument checks whether a function can accept a named argument called "name",
// either directly or via **kwargs.
func acceptsNameArgument(def *build.DefStmt) bool {
	for _, param := range def.Params {
		switch param := param.(type) {
		case *build.Ident:
			if param.Name == "name" {
				return true
			}
		case *build.AssignExpr:
			if ident, ok := param.LHS.(*build.Ident); ok && ident.Name == "name" {
				return true
			}
		case *build.UnaryExpr:
			if param.Op == "**" {
				return true
			}
		}
	}
	return false
}

// fileData represents information about rules and functions extracted from a file
type fileData struct {
	rules     map[string]int
	functions map[string]map[string]funCall
}

// externalDependency is a reference to a symbol defined in another file
type externalDependency struct {
	filename string
	symbol   string
}

// getFunCalls extracts information about functions that are being called from the given function
func getFunCalls(def *build.DefStmt, filename string, externalSymbols map[string]externalDependency) map[string]funCall {
	funCalls := make(map[string]funCall)
	build.Walk(def, func(expr build.Expr, stack []build.Expr) {
		call, ok := expr.(*build.CallExpr)
		if !ok {
			return
		}
		if ident, ok := call.X.(*build.Ident); ok {
			start, _ := call.Span()
			f := funCall{
				function: function{
					filename: filename,
					name:     ident.Name,
				},
				filename:  filename,
				nameAlias: ident.Name,
				line:      start.Line,
				caller:    def.Name,
			}
			if external, ok := externalSymbols[ident.Name]; ok {
				// The function is defined in another file
				f.function = function{
					filename: external.filename,
					name:     external.symbol,
				}
			}
			funCalls[ident.Name] = f
			return
		}
		dot, ok := call.X.(*build.DotExpr)
		if !ok {
			return
		}
		if ident, ok := dot.X.(*build.Ident); !ok || ident.Name != "native" {
			return
		}
		start, _ := dot.Span()
		name := "native." + dot.Name
		funCalls[name] = funCall{
			function: function{
				name:     dot.Name,
				filename: nativeModule,
			},
			filename:  filename,
			nameAlias: name,
			line:      start.Line,
			caller:    def.Name,
		}
	})
	return funCalls
}

// analyzeFile extracts the information about rules and functions defined in the file
func analyzeFile(f *build.File) fileData {
	if f == nil {
		return fileData{}
	}

	// Collect loaded symbols
	externalSymbols := make(map[string]externalDependency)
	for _, stmt := range f.Stmt {
		load, ok := stmt.(*build.LoadStmt)
		if !ok {
			continue
		}
		moduleFile := getPathFromLabel(load.Module.Value, f.Pkg)
		if moduleFile == "" {
			continue
		}
		for i, from := range load.From {
			externalSymbols[load.To[i].Name] = externalDependency{moduleFile, from.Name}
		}
	}

	report := fileData{
		rules:     make(map[string]int),
		functions: make(map[string]map[string]funCall),
	}
	for _, stmt := range f.Stmt {
		switch stmt := stmt.(type) {
		case *build.AssignExpr:
			call, ok := stmt.RHS.(*build.CallExpr)
			if !ok {
				continue
			}
			ident, ok := call.X.(*build.Ident)
			if !ok || ident.Name != "rule" {
				continue
			}
			lhsIdent, ok := stmt.LHS.(*build.Ident)
			if !ok {
				continue
			}
			start, _ := stmt.Span()
			report.rules[lhsIdent.Name] = start.Line
		case *build.DefStmt:
			report.functions[stmt.Name] = getFunCalls(stmt, f.DisplayPath(), externalSymbols)
		default:
			continue
		}
	}
	return report
}

// macroAnalyzer is an object that analyzes the directed graph of functions calling each other,
// loading other files lazily if necessary.
type macroAnalyzer struct {
	fileReader *FileReader
	files      map[string]fileData
	cache      map[function]struct {
		isMacro    bool
		stacktrace []frame
	}
}

// getFileData retrieves a file using the fileReader object and extracts information about functions and rules
// defined in the file.
func (ma macroAnalyzer) getFileData(filename string) fileData {
	if fd, ok := ma.files[filename]; ok {
		return fd
	}
	if ma.fileReader == nil {
		fd := fileData{}
		ma.files[filename] = fd
		return fd
	}
	f := ma.fileReader.GetFile(filename)
	fd := analyzeFile(f)
	ma.files[filename] = fd
	return fd
}

// IsMacro is a public function that checks whether the given function is a macro
func (ma macroAnalyzer) IsMacro(fn function) (bool, []frame) {
	// Keep track of already visited functions to prevent crashing because of infinite recursion
	visited := make(map[function]bool)
	return ma.isMacroPrivate(fn, visited)
}

// isMacroPrivate is the same as IsMacro except that it takes a set of already visited nodes to prevent
// infinite recursion (recursion is not allowed in Starlark but can still crash Buildifier if not
// handled properly)
func (ma macroAnalyzer) isMacroPrivate(fn function, visited map[function]bool) (isMacro bool, stacktrace []frame) {
	if visited[fn] {
		// Don't visit the function again to prevent infinite recursion
		return false, nil
	}
	visited[fn] = true
	defer func() { visited[fn] = false }()

	// Check the cache first
	if cached, ok := ma.cache[fn]; ok {
		return cached.isMacro, cached.stacktrace
	}
	defer func() {
		ma.cache[fn] = struct {
			isMacro    bool
			stacktrace []frame
		}{isMacro, stacktrace}
	}()

	// Check for native rules
	if fn.filename == nativeModule {
		switch fn.name {
		case "glob", "existing_rule", "existing_rules":
			// Not a rule
			return false, nil
		default:
			return true, nil
		}
	}

	fileData := ma.getFileData(fn.filename)
	if line, ok := fileData.rules[fn.name]; ok {
		return true, []frame{ruleDef{
			filename: fn.filename,
			name:     fn.name,
			line:     line,
		}}
	}
	funCalls, ok := fileData.functions[fn.name]
	if !ok {
		return false, nil
	}

	// Prioritize function calls from already loaded files. If some of the function calls are from the same file
	// (or another file that has been loaded already), check them first.
	knownFunCalls := []funCall{}
	newFunCalls := []funCall{}
	for _, fc := range funCalls {
		if _, ok := ma.files[fc.function.filename]; ok || fc.function.filename == nativeModule {
			knownFunCalls = append(knownFunCalls, fc)
		} else {
			newFunCalls = append(newFunCalls, fc)
		}
	}

	for _, fc := range append(knownFunCalls, newFunCalls...) {
		isMacro, stacktrace := ma.isMacroPrivate(fc.function, visited)
		if isMacro {
			return true, append(stacktrace, fc)
		}
	}

	return false, nil
}

func newMacroAnalyzer(fileReader *FileReader) macroAnalyzer {
	return macroAnalyzer{
		fileReader: fileReader,
		files:      make(map[string]fileData),
		cache: make(map[function]struct {
			isMacro    bool
			stacktrace []frame
		}),
	}
}

func unnamedMacroWarning(f *build.File, fileReader *FileReader) []*LinterFinding {
	if f.Type != build.TypeBzl {
		return nil
	}

	macroAnalyzer := newMacroAnalyzer(fileReader)
	macroAnalyzer.files[f.DisplayPath()] = analyzeFile(f)

	findings := []*LinterFinding{}
	for _, stmt := range f.Stmt {
		def, ok := stmt.(*build.DefStmt)
		if !ok {
			continue
		}

		if !ok || acceptsNameArgument(def) {
			continue
		}

		isMacro, stackTrace := macroAnalyzer.IsMacro(function{f.DisplayPath(), def.Name})
		if !isMacro {
			continue
		}
		finding := makeLinterFinding(def, fmt.Sprintf(`Macro function %q doesn't accept a keyword argument "name".

%s
`, def.Name, formatStackTrace(stackTrace)))
		finding.End = def.ColonPos

		findings = append(findings, finding)
	}

	return findings
}
