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

// LintMode is an enum representing a linter mode. Can be either "warn", "fix", or "suggest"
type LintMode int

const (
	// ModeWarn means only warnings should be returned for each finding.
	ModeWarn LintMode = iota
	// ModeFix means that all warnings that can be fixed automatically should be fixed and
	// no warnings should be returned for them.
	ModeFix
	// ModeSuggest means that automatic fixes shouldn't be applied, but instead corresponding
	// suggestions should be attached to all warnings that can be fixed automatically.
	ModeSuggest
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
	Start       int
	End         int
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

// RuleWarningMap lists the warnings that run on a single rule.
// These warnings run only on BUILD files (not bzl files).
var RuleWarningMap = map[string]func(f *build.File, pkg string, expr build.Expr) *Finding{}

// FileWarningMap lists the warnings that run on the whole file.
var FileWarningMap = map[string]func(f *build.File) []*LinterFinding{
	"attr-cfg":            attrConfigurationWarning,
	"attr-license":        attrLicenseWarning,
	"attr-non-empty":      attrNonEmptyWarning,
	"attr-output-default": attrOutputDefaultWarning,
	"attr-single-file":    attrSingleFileWarning,
	"build-args-kwargs":   argsKwargsInBuildFilesWarning,
	"constant-glob":       constantGlobWarning,
	"ctx-actions":         ctxActionsWarning,
	"ctx-args":            contextArgsAPIWarning,
	"duplicated-name":     duplicatedNameWarning,
	"filetype":            fileTypeWarning,
	"git-repository":      nativeGitRepositoryWarning,
	"http-archive":        nativeHTTPArchiveWarning,
	"load":                unusedLoadWarning,
	"native-android":      nativeAndroidRulesWarning,
	"native-build":        nativeInBuildFilesWarning,
	"native-package":      nativePackageWarning,
	"no-effect":           noEffectWarning,
	"output-group":        outputGroupWarning,
	"package-name":        packageNameWarning,
	"positional-args":     RuleWarning(positionalArgumentsWarning),
	"print":               printWarning,
	"redefined-variable":  redefinedVariableWarning,
	"repository-name":     repositoryNameWarning,
	"rule-impl-return":    ruleImplReturnWarning,
	"return-value":        missingReturnValueWarning,
	"uninitialized":       uninitializedVariableWarning,
	"unreachable":         unreachableStatementWarning,
	"unused-variable":     unusedVariableWarning,
}

// LegacyFileWarningMap lists the warnings that run on the whole file with legacy interface.
var LegacyFileWarningMap = map[string]func(f *build.File, fix bool) []*Finding{
	"confusing-name":            confusingNameWarning,
	"depset-iteration":          depsetIterationWarning,
	"depset-union":              depsetUnionWarning,
	"dict-concatenation":        dictionaryConcatenationWarning,
	"function-docstring":        functionDocstringWarning,
	"function-docstring-header": functionDocstringHeaderWarning,
	"function-docstring-args":   functionDocstringArgsWarning,
	"function-docstring-return": functionDocstringReturnWarning,
	"integer-division":          integerDivisionWarning,
	"load-on-top":               loadOnTopWarning,
	"module-docstring":          moduleDocstringWarning,
	"name-conventions":          nameConventionsWarning,
	"out-of-order-load":         outOfOrderLoadWarning,
	"package-on-top":            packageOnTopWarning,
	"same-origin-load":          sameOriginLoadWarning,
	"string-iteration":          stringIterationWarning,
	"unsorted-dict-items":       unsortedDictItemsWarning,
}

// nonDefaultWarnings contains warnings that are enabled by default because they're not applicable
// for all files and cause too much diff noise when applied.
var nonDefaultWarnings = map[string]bool{
	"out-of-order-load":   true, // load statements should be sorted by their labels
	"unsorted-dict-items": true, // dict items should be sorted
}

// RuleWarning is a wrapper that converts a per-rule function to a per-file function. It also doesn't
// run on .bzl of default files.
func RuleWarning(ruleWarning func(call *build.CallExpr) []*LinterFinding) func(f *build.File) []*LinterFinding {
	return func(f *build.File) []*LinterFinding {
		if f.Type != build.TypeBuild && f.Type != build.TypeWorkspace {
			return nil
		}
		findings := []*LinterFinding{}
		for _, stmt := range f.Stmt {
			switch stmt := stmt.(type) {
			case *build.CallExpr:
				findings = append(findings, ruleWarning(stmt)...)
			case *build.Comprehension:
				// Rules are often called within list comprehensions, e.g. [my_rule(foo) for foo in bar]
				if call, ok := stmt.Body.(*build.CallExpr); ok {
					findings = append(findings, ruleWarning(call)...)
				}
			}
		}
		return findings
	}
}

// DisabledWarning checks if the warning was disabled by a comment.
// The comment format is buildozer: disable=<warning>
func DisabledWarning(f *build.File, findingLine int, warning string) bool {
	format := "buildozer: disable=" + warning

	for _, stmt := range f.Stmt {
		if stmt == nil {
			continue
		}
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
func FileWarnings(f *build.File, pkg string, enabledWarnings []string, formatted *[]byte, mode LintMode) []*Finding {
	findings := []*Finding{}

	// Sort the warnings to make sure they're applied in the same determined order
	// Make a local copy first to avoid race conditions
	warnings := append([]string{}, enabledWarnings...)
	sort.Strings(warnings)

	// If suggestions are requested and formatted file is not provided, format it to compare modified versions with
	if mode == ModeSuggest && formatted == nil {
		contents := build.Format(f)
		formatted = &contents
	}

	for _, warn := range warnings {
		if fct, ok := FileWarningMap[warn]; ok {
			findings = append(findings, runFileWarningsFunction(warn, f, fct, formatted, mode)...)
		} else if fct, ok := LegacyFileWarningMap[warn]; ok {
			for _, w := range fct(f, mode == ModeFix) {
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
func runFileWarningsFunction(category string, f *build.File, fct func(f *build.File) []*LinterFinding, formatted *[]byte, mode LintMode) []*Finding {
	findings := []*Finding{}
	for _, w := range fct(f) {
		if !DisabledWarning(f, w.Start.Line, category) {
			finding := makeFinding(f, w.Start, w.End, category, w.Message, true, nil)
			if len(w.Replacement) > 0 {
				// An automatic fix exists
				switch mode {
				case ModeFix:
					// Apply the fix and discard the finding
					for _, r := range w.Replacement {
						*r.Old = r.New
					}
					finding = nil
				case ModeSuggest:
					// Apply the fix, calculate the diff and roll back the fix
					newContents := formatWithFix(f, &w.Replacement)

					start, end, replacement := calculateDifference(formatted, &newContents)
					finding.Replacement = &Replacement{
						Description: w.Message,
						Start:       start,
						End:         end,
						Content:     replacement,
					}
				}
			}
			if finding != nil {
				findings = append(findings, finding)
			}
		}
	}
	return findings
}

// formatWithFix applies a fix, formats a file, and rolls back the fix
func formatWithFix(f *build.File, replacements *[]LinterReplacement) []byte {
	for i := range *replacements {
		r := (*replacements)[i]
		old := *r.Old
		*r.Old = r.New
		defer func() { *r.Old = old }()
	}

	return build.Format(f)
}

// calculateDifference compares two file contents and returns a replacement in the form of
// a 3-tuple (byte from, byte to (non inclusive), a string to replace with).
func calculateDifference(old, new *[]byte) (start, end int, replacement string) {
	commonPrefix := 0 // length of the common prefix
	for i, b := range *old {
		if i >= len(*new) || b != (*new)[i] {
			break
		}
		commonPrefix++
	}

	commonSuffix := 0 // length of the common suffix
	for i := range *old {
		b := (*old)[len(*old)-1-i]
		if i >= len(*new) || b != (*new)[len(*new)-1-i] {
			break
		}
		commonSuffix++
	}

	// In some cases common suffix and prefix can overlap. E.g. consider the following case:
	//   old = "abc"
	//   new = "abdbc"
	// In this case the common prefix is "ab" and the common suffix is "bc".
	// If they overlap, just shorten the suffix so that they don't.
	// The new suffix will be just "c".
	if commonPrefix+commonSuffix > len(*old) {
		commonSuffix = len(*old) - commonPrefix
	}
	if commonPrefix+commonSuffix > len(*new) {
		commonSuffix = len(*new) - commonPrefix
	}
	return commonPrefix, len(*old) - commonSuffix, string((*new)[commonPrefix:(len(*new) - commonSuffix)])
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
	warnings := FileWarnings(f, pkg, enabledWarnings, nil, ModeFix)
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
