// Package warn implements functions that generate warnings for BUILD files.
package warn

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
)

// LinterFinding is a low-level warning reported by single linter/fixer functions.
type LinterFinding struct {
	Start       build.Position
	End         build.Position
	Message     string
	Replacement []LinterReplacement
}

// LinterReplacement is a low-level object returned by single fixer functions.
type LinterReplacement struct {
	Old *build.Expr
	New build.Expr
}

// A Finding is a warning reported by the analyzer. It may contain an optional suggested fix.
type Finding struct {
	File        *build.File
	Start       build.Position
	End         build.Position
	Category    string
	Message     string
	URL         string
	Actionable  bool
	Replacement *Replacement
}

// A Replacement is a suggested fix. Text between Start and End should be replaced with Content.
type Replacement struct {
	Description string
	Start       build.Position
	End         build.Position
	Content     string
}

func docURL(cat string) string {
	return "https://github.com/bazelbuild/buildtools/blob/master/WARNINGS.md#" + cat
}

// makeFinding creates a Finding object
func makeFinding(f *build.File, start, end build.Position, cat string, msg string, actionable bool, fix *Replacement) *Finding {
	return &Finding{
		File:        f,
		Start:       start,
		End:         end,
		Category:    cat,
		URL:         docURL(cat),
		Message:     msg,
		Actionable:  actionable,
		Replacement: fix,
	}
}

// makeLinterFinding creates a LinterFinding object
func makeLinterFinding(node build.Expr, message string, replacement ...LinterReplacement) *LinterFinding {
	start, end := node.Span()
	return &LinterFinding{
		Start:       start,
		End:         end,
		Message:     message,
		Replacement: replacement,
	}
}

// MakeFix creates a Replacement object
func MakeFix(f *build.File, desc string, start build.Position, end build.Position, newContent string) *Replacement {
	return &Replacement{
		Description: desc,
		Start:       start,
		End:         end,
		Content:     newContent,
	}
}

// RuleWarningMap lists the warnings that run on a single rule.
// These warnings run only on BUILD files (not bzl files).
var RuleWarningMap = map[string]func(f *build.File, pkg string, expr build.Expr) *Finding{
	"positional-args": positionalArgumentsWarning,
}

// FileWarningMap lists the warnings that run on the whole file.
var FileWarningMap = map[string]func(f *build.File) []*LinterFinding{
	"attr-cfg":     attrConfigurationWarning,
	"attr-license": attrLicenseWarning,
}

// LegacyFileWarningMap lists the warnings that run on the whole file with legacy interface.
var LegacyFileWarningMap = map[string]func(f *build.File, fix bool) []*Finding{
	"attr-non-empty":            attrNonEmptyWarning,
	"attr-output-default":       attrOutputDefaultWarning,
	"attr-single-file":          attrSingleFileWarning,
	"build-args-kwargs":         argsKwargsInBuildFilesWarning,
	"confusing-name":            confusingNameWarning,
	"constant-glob":             constantGlobWarning,
	"ctx-actions":               ctxActionsWarning,
	"ctx-args":                  contextArgsAPIWarning,
	"depset-iteration":          depsetIterationWarning,
	"depset-union":              depsetUnionWarning,
	"dict-concatenation":        dictionaryConcatenationWarning,
	"duplicated-name":           duplicatedNameWarning,
	"filetype":                  fileTypeWarning,
	"function-docstring":        functionDocstringWarning,
	"function-docstring-header": functionDocstringHeaderWarning,
	"function-docstring-args":   functionDocstringArgsWarning,
	"function-docstring-return": functionDocstringReturnWarning,
	"git-repository":            nativeGitRepositoryWarning,
	"http-archive":              nativeHTTPArchiveWarning,
	"integer-division":          integerDivisionWarning,
	"load":                      unusedLoadWarning,
	"load-on-top":               loadOnTopWarning,
	"return-value":              missingReturnValueWarning,
	"module-docstring":          moduleDocstringWarning,
	"name-conventions":          nameConventionsWarning,
	"native-android":            nativeAndroidRulesWarning,
	"native-build":              nativeInBuildFilesWarning,
	"native-package":            nativePackageWarning,
	"no-effect":                 noEffectWarning,
	"out-of-order-load":         outOfOrderLoadWarning,
	"output-group":              outputGroupWarning,
	"package-name":              packageNameWarning,
	"package-on-top":            packageOnTopWarning,
	"redefined-variable":        redefinedVariableWarning,
	"repository-name":           repositoryNameWarning,
	"rule-impl-return":          ruleImplReturnWarning,
	"same-origin-load":          sameOriginLoadWarning,
	"string-iteration":          stringIterationWarning,
	"uninitialized":             uninitializedVariableWarning,
	"unreachable":               unreachableStatementWarning,
	"unsorted-dict-items":       unsortedDictItemsWarning,
	"unused-variable":           unusedVariableWarning,
}

// nonDefaultWarnings contains warnings that are enabled by default because they're not applicable
// for all files and cause too much diff noise when applied.
var nonDefaultWarnings = map[string]bool{
	"out-of-order-load":   true, // load statements should be sorted by their labels
	"unsorted-dict-items": true, // dict items should be sorted
}

// DisabledWarning checks if the warning was disabled by a comment.
// The comment format is buildozer: disable=<warning>
func DisabledWarning(f *build.File, findingLine int, warning string) bool {
	format := "buildozer: disable=" + warning

	for _, stmt := range f.Stmt {
		stmtStart, _ := stmt.Span()
		if stmtStart.Line == findingLine {
			// Is this specific line disabled?
			if edit.ContainsComments(stmt, format) {
				return true
			}
		}
		// Check comments within a rule
		rule, ok := stmt.(*build.CallExpr)
		if ok {
			for _, stmt := range rule.List {
				stmtStart, _ := stmt.Span()
				if stmtStart.Line != findingLine {
					continue
				}
				// Is the whole rule or this specific line as a comment
				// to disable this warning?
				if edit.ContainsComments(rule, format) ||
					edit.ContainsComments(stmt, format) {
					return true
				}
			}
		}
		// Check comments within a load statement
		load, ok := stmt.(*build.LoadStmt)
		if ok {
			loadHasComment := edit.ContainsComments(load, format)
			module := load.Module
			if module.Start.Line == findingLine {
				if edit.ContainsComments(module, format) || loadHasComment {
					return true
				}
			}
			for i, to := range load.To {
				from := load.From[i]
				if to.NamePos.Line == findingLine || from.NamePos.Line == findingLine {
					if edit.ContainsComments(to, format) || edit.ContainsComments(from, format) || loadHasComment {
						return true
					}
				}
			}
		}
	}

	return false
}

// FileWarnings returns a list of all warnings found in the file.
func FileWarnings(f *build.File, pkg string, enabledWarnings []string, fix bool) []*Finding {
	findings := []*Finding{}

	// Sort the warnings to make sure they're applied in the same determined order
	// Make a local copy first to avoid race conditions
	warnings := append([]string{}, enabledWarnings...)
	sort.Strings(warnings)

	for _, warn := range warnings {
		if fct, ok := FileWarningMap[warn]; ok {
			findings = append(findings, runFileWarningsFunction(warn, f, fct, fix)...)
		} else if fct, ok := LegacyFileWarningMap[warn]; ok {
			for _, w := range fct(f, fix) {
				if !DisabledWarning(f, w.Start.Line, warn) {
					findings = append(findings, w)
				}
			}
		} else if fct, ok := RuleWarningMap[warn]; ok {
			findings = append(findings, runRuleWarningsFunction(warn, pkg, f, fct)...)
		} else {
			log.Fatalf("unexpected warning %q", warn)
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Start.Line < findings[j].Start.Line })
	return findings
}

// runFileWarningsFunction runs a linter/fixer function over a file and applies the fixes conditionally
func runFileWarningsFunction(category string, f *build.File, fct func(f *build.File) []*LinterFinding, fix bool) []*Finding {
	findings := []*Finding{}
	for _, w := range fct(f) {
		if !DisabledWarning(f, w.Start.Line, category) {
			if fix && len(w.Replacement) > 0 {
				for _, r := range w.Replacement {
					*r.Old = r.New
				}
			} else {
				findings = append(findings, makeFinding(f, w.Start, w.End, category, w.Message, true, nil))
			}
		}
	}
	return findings
}

// runRuleWarningsFunction runs a linter/fixer function over each rule in file
func runRuleWarningsFunction(category, pkg string, f *build.File, fct func(f *build.File, pkg string, expr build.Expr) *Finding) []*Finding {
	if f.Type != build.TypeBuild && f.Type != build.TypeWorkspace {
		return nil
	}
	findings := []*Finding{}
	for _, stmt := range f.Stmt {
		if w := fct(f, pkg, stmt); w != nil {
			if !DisabledWarning(f, w.Start.Line, category) {
				findings = append(findings, w)
			}
		}
	}
	return findings
}

// FixWarnings fixes all warnings that can be fixed automatically.
func FixWarnings(f *build.File, pkg string, enabledWarnings []string, verbose bool) {
	warnings := FileWarnings(f, pkg, enabledWarnings, true)
	if verbose {
		fmt.Fprintf(os.Stderr, "%s: applied fixes, %d warnings left\n",
			f.DisplayPath(),
			len(warnings))
	}
}

func collectAllWarnings() []string {
	var result []string
	// Collect list of all warnings.
	for k := range FileWarningMap {
		result = append(result, k)
	}
	for k := range LegacyFileWarningMap {
		result = append(result, k)
	}
	for k := range RuleWarningMap {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// AllWarnings is the list of all available warnings.
var AllWarnings = collectAllWarnings()

func collectDefaultWarnings() []string {
	warnings := []string{}
	for _, warning := range AllWarnings {
		if !nonDefaultWarnings[warning] {
			warnings = append(warnings, warning)
		}
	}
	return warnings
}

// DefaultWarnings is the list of all warnings that should be used inside google3
var DefaultWarnings = collectDefaultWarnings()
