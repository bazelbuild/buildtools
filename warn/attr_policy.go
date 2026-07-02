package warn

import (
	"path"
	"strings"

	"github.com/bazelbuild/buildtools/labels"
)

// AttrPolicyConstraintFamily identifies which constraint fields apply to a policy rule.
type AttrPolicyConstraintFamily int

const (
	AttrPolicyScalarFamily AttrPolicyConstraintFamily = iota
	AttrPolicyListFamily
	AttrPolicyDictFamily
	AttrPolicyNumericFamily
)

// AttrPolicyAllowlistKind is a compiled allow-list pattern kind.
type AttrPolicyAllowlistKind int

const (
	AttrPolicyAllowAll AttrPolicyAllowlistKind = iota
	AttrPolicyAllowExact
	AttrPolicyAllowPackageAll
	AttrPolicyAllowRecursive
)

// AttrPolicyAllowlistPattern is a compiled allow-list entry.
type AttrPolicyAllowlistPattern struct {
	Kind   AttrPolicyAllowlistKind
	Pkg    string
	Target string
}

// AttrPolicyRuleCompiled is a compiled attribute policy rule.
type AttrPolicyRuleCompiled struct {
	Name     string
	RuleKinds []string
	Attr     string
	Family   AttrPolicyConstraintFamily

	ForbidValues       []string
	RequireValues      []string
	ForbidListItems    []string
	RequireListItems   []string
	ForbidDictEntries  map[string]string
	RequireDictEntries map[string]string
	ForbidDictKeys     []string
	MinValue           *int
	MaxValue           *int

	Required     bool
	Allowlist    []AttrPolicyAllowlistPattern
	Suppressible bool
	Message      string
}

// AttrPolicyConfig is process-global policy, set from buildifier config before linting.
var AttrPolicyConfig []AttrPolicyRuleCompiled

// SetAttrPolicy replaces the active attribute policy rules.
func SetAttrPolicy(rules []AttrPolicyRuleCompiled) {
	AttrPolicyConfig = rules
}

func matchesRuleKind(globs []string, kind string) bool {
	if len(globs) == 0 {
		return true
	}
	for _, g := range globs {
		if matched, err := path.Match(g, kind); err == nil && matched {
			return true
		}
	}
	return false
}

func allowlistMatches(patterns []AttrPolicyAllowlistPattern, label, pkg string) bool {
	for _, p := range patterns {
		if allowlistPatternMatches(p, label, pkg) {
			return true
		}
	}
	return false
}

func allowlistPatternMatches(p AttrPolicyAllowlistPattern, label, pkg string) bool {
	switch p.Kind {
	case AttrPolicyAllowAll:
		return true
	case AttrPolicyAllowRecursive:
		l := labels.Parse(label)
		return l.Package == p.Pkg || strings.HasPrefix(l.Package, p.Pkg+"/")
	case AttrPolicyAllowPackageAll:
		return labels.Parse(label).Package == p.Pkg
	case AttrPolicyAllowExact:
		want := labels.Label{Package: p.Pkg, Target: p.Target}.Format()
		return labels.Equal(label, want, pkg)
	default:
		return false
	}
}
