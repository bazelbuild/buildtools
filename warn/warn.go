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
var FileWarningMap = map[string]func(f *build.File, fix bool) []*Finding{
	"attr-cfg":            attrConfigurationWarning,
	"attr-license":        attrLicenseWarning,
	"attr-non-empty":      attrNonEmptyWarning,
	"attr-output-default": attrOutputDefaultWarning,
	"attr-single-file":    attrSingleFileWarning,
	"confusing-name":      confusingNameWarning,
	"constant-glob":       constantGlobWarning,
	"ctx-actions":         ctxActionsWarning,
	"ctx-args":            contextArgsAPIWarning,
	"depset-iteration":    depsetIterationWarning,
	"depset-union":        depsetUnionWarning,
	"dict-concatenation":  dictionaryConcatenationWarning,
	"duplicated-name":     duplicatedNameWarning,
	"filetype":            fileTypeWarning,
	"function-docstring":  functionDocstringWarning,
	"git-repository":      nativeGitRepositoryWarning,
	"http-archive":        nativeHTTPArchiveWarning,
	"integer-division":    integerDivisionWarning,
	"load":                unusedLoadWarning,
	"load-on-top":         loadOnTopWarning,
	"return-value":        missingReturnValueWarning,
	"module-docstring":    moduleDocstringWarning,
	"name-conventions":    nameConventionsWarning,
	"native-build":        nativeInBuildFilesWarning,
	"native-package":      nativePackageWarning,
	"no-effect":           noEffectWarning,
	"out-of-order-load":   outOfOrderLoadWarning,
	"output-group":        outputGroupWarning,
	"package-name":        packageNameWarning,
	"package-on-top":      packageOnTopWarning,
	"redefined-variable":  redefinedVariableWarning,
	"repository-name":     repositoryNameWarning,
	"rule-impl-return":    ruleImplReturnWarning,
	"same-origin-load":    sameOriginLoadWarning,
	"string-iteration":    stringIterationWarning,
	"uninitialized":       uninitializedVariableWarning,
	"unreachable":         unreachableStatementWarning,
	"unsorted-dict-items": unsortedDictItemsWarning,
	"unused-variable":     unusedVariableWarning,
}

// nonDefaultWarnings contains warnings that are enabled by default because they're not applicable
// for all files and cause too much diff noise when applied.
var nonDefaultWarnings = map[string]bool{
	"out-of-order-load":   true, // load statements should be sorted by their labels
	"unsorted-dict-items": true, // dict items should be sorted
}

// DisabledWarning checks if the warning was disabled by a comment.
// The comment format is buildozer: disable=<warning>
func DisabledWarning(f *build.File, finding *Finding, warning string) bool {
	format := "buildozer: disable=" + warning
	findingLine := finding.Start.Line

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
	for _, warn := range enabledWarnings {
		if fct, ok := FileWarningMap[warn]; ok {
			for _, w := range fct(f, fix) {
				if !DisabledWarning(f, w, warn) {
					findings = append(findings, w)
				}
			}
		} else {
			fn := RuleWarningMap[warn]
			if fn == nil {
				log.Fatalf("unexpected warning %q", warn)
			}
			if f.Type == build.TypeDefault {
				continue
			}
			for _, stmt := range f.Stmt {
				if w := fn(f, pkg, stmt); w != nil {
					if !DisabledWarning(f, w, warn) {
						findings = append(findings, w)
					}
				}
			}
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Start.Line < findings[j].Start.Line })
	return findings
}

// PrintWarnings prints the list of warnings returned from calling FileWarnings.
// Actionable warnings list their link in parens, inactionable warnings list
// their link in square brackets.
func PrintWarnings(f *build.File, warnings []*Finding, showReplacements bool) {
	for _, w := range warnings {
		formatString := "%s:%d: %s: %s (%s)"
		if !w.Actionable {
			formatString = "%s:%d: %s: %s [%s]"
		}
		fmt.Fprintf(os.Stderr, formatString,
			w.File.DisplayPath(),
			w.Start.Line,
			w.Category,
			w.Message,
			w.URL)
		if showReplacements && w.Replacement != nil {
			r := w.Replacement
			fmt.Fprintf(os.Stderr, " [%d..%d): %s\n",
				r.Start.Byte,
				r.End.Byte,
				r.Content)
		} else {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
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
