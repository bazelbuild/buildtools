/*
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/bazelbuild/buildtools/add_dep/adddep"
	"github.com/bazelbuild/buildtools/add_dep/bazel"
	"github.com/bazelbuild/buildtools/add_dep/color"
	"github.com/bazelbuild/buildtools/add_dep/query"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
	"github.com/bazelbuild/buildtools/file"
)

var (
	buildifier = flag.String("buildifier", "buildifier", "path to buildifier binary")
	comment    = flag.String("comment", "", "comment to add to the end of the line e.g. 'buildcleaner: keep'")

	attribute = flag.String("attribute", "deps", "deprecated; do not use")
)

type aspectToAdd struct {
	after bazel.Label
	kind  string
	label bazel.Label
}

func main() {
	flag.Parse()

	fixlabel, deps, aspects, err := parseArgs(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		fmt.Fprintf(os.Stderr, "usage: add_dep (\"<dep label>( <aspect name>)?\")+ <label to fix>\n")
		os.Exit(1)
	}
	if err := run(context.Background(), fixlabel, deps, aspects); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.Red("FAILED:"), err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (bazel.Label, []bazel.Label, map[bazel.Label]string, error) {
	if len(args) < 2 {
		return "", nil, nil, errors.New("need at least two args")
	}

	aspects := make(map[bazel.Label]string)
	var deps []bazel.Label
	for _, a := range args[:len(args)-1] {
		fields := strings.Fields(a)
		if len(fields) > 2 {
			return "", nil, nil, fmt.Errorf("invalid label: %s", a)
		}
		label, err := bazel.ParseAbsoluteLabel(fields[0])
		if err != nil {
			return "", nil, nil, fmt.Errorf("could not parse label %s: %v", a, err)
		}
		if len(fields) == 2 {
			aspects[label] = fields[1]
		}
		deps = append(deps, label)
	}

	fixlabel, err := bazel.ParseAbsoluteLabel(args[len(args)-1])
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not parse label %s: %v", args[len(args)-1], err)
	}
	return fixlabel, deps, aspects, nil
}

func run(ctx context.Context, fixlabel bazel.Label, newdeps []bazel.Label, aspects map[bazel.Label]string) error {
	now := time.Now()
	// fixlabel is trivally visible to itself, but retrieving the rule this way
	// avoids another query.
	visible, err := getVisibility(ctx, fixlabel, append(newdeps, fixlabel))
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}
	fixrule, ok := visible[fixlabel]
	if !ok {
		return fmt.Errorf("could not query %s", fixlabel)
	}
	if fixrule.RuleClass() == "proto_library" {
		return errors.New("adding deps to proto_library rules is not currently supported")
	}
	failed := 0
	var pendingAdds []bazel.Label
	seen := make(map[bazel.Label]bool)
	aspectsToAdd := make(map[bazel.Label]aspectToAdd)
	for _, dep := range newdeps {

		aspect, _ := aspects[dep]
		toAdd, err := resolve(ctx, visible, fixrule, dep, aspect, aspectsToAdd)
		if err != nil {
			fmt.Printf("%s %s: %v\n", color.Red("FAILED:"), dep, err)
			failed++
			continue
		}
		if r, ok := visible[toAdd]; ok && r.RuleClass() == "proto_library" {
			return fmt.Errorf("unexpected proto_library dependency %s", toAdd)
		}

		if fixrule.HasDep(toAdd) {
			fmt.Printf("%s %s (already present)\n", color.Yellow("SKIPPING:"), toAdd)
			continue
		}
		if _, ok := seen[toAdd]; ok {
			fmt.Printf("%s %s (already added)\n", color.Yellow("SKIPPING:"), toAdd)
			continue
		}
		fmt.Printf("%s %s", color.Green("ADDING:"), toAdd)
		if toAdd != dep {
			fmt.Printf(" (resolved from %s)\n", dep)
		} else {
			fmt.Println()
		}
		pendingAdds = append(pendingAdds, toAdd)
		seen[toAdd] = true
	}
	if err = addDeps(ctx, fixrule, pendingAdds); err != nil {
		return fmt.Errorf("could not add dependencies to %s: %v", fixlabel, err)
	}
	// Run buildifier on the edited file.
	if path, _, err := fixrule.Location(); err == nil {
		if err := exec.Command(*buildifier, path).Run(); err != nil {
			return fmt.Errorf("failed to buildify %s: %v", path, err)
		}
	}
	for _, a := range aspectsToAdd {
		err := addRuleAfter(ctx, a.after, a.kind, a.label)
		if err != nil {
			return err
		}
	}
	elapsed := time.Since(now).Nanoseconds() / int64(time.Millisecond)
	fmt.Printf("finished in %dms [added %d/%d dep(s)]\n\n", elapsed, len(pendingAdds), len(newdeps))
	if failed > 0 {
		return errors.New("not all dependencies could be added")
	}
	return nil
}

// resolveAspect resolves a label to add for the given dependency and aspect.
func resolveAspect(ctx context.Context, visible map[bazel.Label]query.Rule, from query.Rule, to bazel.Label, aspect string, aspectsToAdd map[bazel.Label]aspectToAdd) (bazel.Label, error) {
	// Look for an existing j_p_l wrapper rule for the dependency in the same package.
	q := adddep.LangProtoLibraryQuery(aspect, to)
	candidates, err := query.Query(ctx, q, true)
	if err != nil {
		return "", err
	}
	pkg, name := to.Split()
	newname, err := adddep.LangProtoLibraryName(name, aspect)
	if err != nil {
		return "", err
	}
	preferredLabel, err := bazel.ParseAbsoluteLabel("//" + pkg + ":" + newname)
	if err != nil {
		return "", err
	}
	if len(candidates) > 0 {
		// If there are multiple wrapper rules, prefer one that follows the naming convention.
		var names []string
		for _, c := range candidates {
			if !c.HasInput(to) {
				continue
			}
			if c.Label == preferredLabel {
				return c.Label, nil
			}
			if c.AvoidDep() {
				continue
			}
			names = append(names, string(c.Label))
		}
		if len(names) > 0 {
			if len(names) > 1 {
				sort.Strings(names)
				fmt.Printf("found multiple candidates for %s: %s\n", to, strings.Join(names, ", "))
			}
			return bazel.Label(names[0]), nil
		}
	}

	// We didn't find any existing wrapper rules, so create a new one.
	rule, err := query.Query(ctx, string(to), false)
	if err != nil {
		return "", err
	}
	if len(rule) != 1 {
		return "", fmt.Errorf("expected exactly one rule for %s, saw %d", to, len(rule))
	}
	aspectsToAdd[preferredLabel] = aspectToAdd{
		after: rule[to].Label,
		kind:  aspect,
		label: preferredLabel,
	}
	return preferredLabel, nil
}

// resolve resolves a transitive dependency to be added to the actual label we
// want to add as a dep, including handling of: exports, visibility, avoid_dep, etc.
func resolve(ctx context.Context, visible map[bazel.Label]query.Rule, from query.Rule, to bazel.Label, aspect string, aspectsToAdd map[bazel.Label]aspectToAdd) (bazel.Label, error) {
	if aspect != "" {
		return resolveAspect(ctx, visible, from, to, aspect, aspectsToAdd)
	}

	if err := isValidDep(visible, to); err == nil {
		return to, nil
	}

	// Visible exports of invisible targets will usually be found on the path from
	// the target being fixed to the dep (especially for third_party).
	qs := fmt.Sprintf("somepath(%s, %s)", from.Label, to)
	rules, err := query.Query(ctx, qs, true)
	if err != nil {
		return "", err
	}
	if len(rules) == 0 {
		return "", errors.New("target is not visible, and is not an existing transitive dependency")
	}
	if label, err := resolveExports(ctx, from, to, rules); err == nil {
		return label, nil
	}

	// If the somepath query didn't turn up a visible target, do a last-ditch search
	// for visible targets in the same package as the dependency.
	qs = fmt.Sprintf("siblings(%s)", to)
	rules, err = query.Query(ctx, qs, true)
	if err != nil {
		return "", err
	}
	if label, err := resolveExports(ctx, from, to, rules); err == nil {
		return label, nil
	}

	// If we couldn't find an alternative, return the error for the original dep.
	return "", isValidDep(visible, to)
}

// resolveExports walks export back-edges in the given universe
// looking for a dep that is visible and not tagged 'avoid_dep'.
func resolveExports(ctx context.Context, from query.Rule, to bazel.Label, universe map[bazel.Label]query.Rule) (bazel.Label, error) {
	backedges := make(map[bazel.Label][]query.Rule)
	for _, r := range universe {
		for _, input := range r.Inputs() {
			if _, ok := universe[input]; ok && r.IsExport(input) {
				backedges[input] = append(backedges[input], r)
			}
		}
	}

	// Find all targets in the query results that are reachable from the dep via export back-edges,
	// and then query to see which of those candidates are visible.
	reachable := make(map[bazel.Label]bool)
	rule, ok := universe[to]
	if !ok {
		log.Fatalf("universe was missing: %s", to)
	}
	search(rule, backedges, reachable, func(r bazel.Label) bool {
		return false
	})
	var labels []bazel.Label
	for k := range reachable {
		labels = append(labels, k)
	}
	visible, err := getVisibility(ctx, from.Label, labels)
	if err != nil {
		return "", fmt.Errorf("getting visibility failed: %v", err)
	}

	// Find a target that is reachable via exports and otherwise valid.
	if label, ok := search(rule, backedges, make(map[bazel.Label]bool), func(label bazel.Label) bool {
		return isValidDep(visible, label) == nil
	}); ok {
		return label, nil
	}
	return "", errors.New("no deps found")
}

// search performs a depth-first search starting at the given rule.
func search(r query.Rule, backedges map[bazel.Label][]query.Rule, seen map[bazel.Label]bool, p func(bazel.Label) bool) (bazel.Label, bool) {
	if p(r.Label) {
		return r.Label, true
	}
	if seen[r.Label] {
		return "", false
	}
	seen[r.Label] = true
	if preds, ok := backedges[r.Label]; ok {
		for _, pred := range preds {
			if label, ok := search(pred, backedges, seen, p); ok {
				return label, true
			}
		}
	}
	return "", false
}

func isValidDep(visible map[bazel.Label]query.Rule, label bazel.Label) error {
	r, vis := visible[label]
	if !vis {
		return errors.New("target is not visible")
	}
	if r.AvoidDep() {
		return errors.New("target has 'avoid_dep' tag")
	}
	switch r.RuleClass() {
	case "proto_library", "java_binary":
		return fmt.Errorf("%s cannot be used as a dependency", r.RuleClass())
	}
	return nil
}

func getVisibility(ctx context.Context, from bazel.Label, labels []bazel.Label) (map[bazel.Label]query.Rule, error) {
	if len(labels) == 0 {
		return map[bazel.Label]query.Rule{}, nil
	}
	var labelStrings []string
	for _, label := range labels {
		labelStrings = append(labelStrings, string(label))
	}
	q := fmt.Sprintf("visible(%s, %s)", from, strings.Join(labelStrings, " + "))
	return query.Query(ctx, q, true)
}

func addDeps(ctx context.Context, r query.Rule, deps []bazel.Label) error {
	f, line, err := r.Location()
	if err != nil {
		return err
	}

	contents, b, err := readBuildFile(ctx, f)
	if err != nil {
		return err
	}
	rule := b.RuleAt(line)
	if rule == nil {
		return fmt.Errorf("%s: unable to find rule at line %d", f, line)
	}

	// We've already checked for an existing dep against the bazel query output,
	// but the dep may be syntactically present even if it's absent in the configured
	// target, e.g. due to a broken macro definition that doesn't pass the deps
	// attribute through to the underlying rule.
	pkg, _ := r.Label.Split()
	for _, dep := range deps {
		for _, a := range rule.AttrStrings(*attribute) {
			l, err := bazel.ParseRelativeLabel(pkg, a)
			if err == nil && l == dep {
				return fmt.Errorf("dep %s was already present", dep)
			}
		}
		short := edit.ShortenLabel(string(dep), pkg)
		expr := &build.StringExpr{Value: short}
		if len(*comment) > 0 {
			tok := fmt.Sprintf("# %s", strings.TrimSpace(*comment))
			expr.Comments.Suffix = []build.Comment{{Token: tok}}
		}
		edit.AddValueToListAttribute(rule, *attribute, pkg, expr, nil)
	}
	return writeBuildFile(ctx, contents, b, r.Label)
}

// addRuleAfter adds a new build immediately after the given rule, and with the given
// rule kind and label.
func addRuleAfter(ctx context.Context, after bazel.Label, kind string, label bazel.Label) error {
	q, err := query.Query(ctx, string(after), true)
	if err != nil {
		return err
	}
	r := q[after]
	f, line, err := r.Location()
	if err != nil {
		return err
	}
	contents, b, err := readBuildFile(ctx, f)
	if err != nil {
		return err
	}

	_, name := label.Split()
	if edit.FindRuleByName(b, name) != nil {
		return fmt.Errorf("%s already exists", label)
	}

	call := &build.CallExpr{X: &build.Ident{Name: kind}}
	rule := &build.Rule{Call: call, ImplicitName: ""}
	rule.SetAttr("name", &build.StringExpr{Value: name})
	_, short := r.Label.Split()
	edit.AddValueToListAttribute(rule, "deps", "", &build.StringExpr{Value: ":" + short}, nil)

	insertionIndex := -1
	for i, s := range b.Stmt {
		start, end := s.Span()
		if start.Line <= line && line <= end.Line {
			insertionIndex = i
			break
		}
	}
	if insertionIndex == -1 {
		return fmt.Errorf("unable to locate rule for %s", r.Label)
	}
	b.Stmt = edit.InsertAfter(insertionIndex, b.Stmt, call)
	if load, ok := adddep.LangProtoLibraryLoad(kind); ok {
		b.Stmt = edit.InsertLoad(b.Stmt, load, []string{kind}, []string{kind})
	}
	return writeBuildFile(ctx, contents, b, label)
}

// readBuildFile reads in and parses the build file at the given path.
func readBuildFile(ctx context.Context, f string) ([]byte, *build.File, error) {
	contents, err := os.ReadFile(f)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: unable to read build file: %v", f, err)
	}
	b, err := build.Parse(f, contents)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: unable to parse build file: %v", f, err)
	}
	return contents, b, nil
}

// writeBuildFile writes out the given build file, and does bazel query to sanity check that
// the resultant build file is still well formed. If the bazel query fails, the original
// contents of the build file are restored.
func writeBuildFile(ctx context.Context, originalContents []byte, b *build.File, label bazel.Label) error {
	stat, err := os.Stat(b.Path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %v", b.Path, err)
	}
	// Call g4 edit on read-only files.
	if stat.Mode()&0222 == 0 {
		if err := exec.Command("g4", "edit", b.Path).Run(); err != nil {
			return fmt.Errorf("failed to open %s for editing: %v", b.Path, err)
		}
	}
	if err := file.WriteFile(b.Path, build.Format(b)); err != nil {
		return fmt.Errorf("failed to write BUILD file: %v", err)
	}
	// Run a post-edit bazel query as a sanity check, and revert the change if the query fails.
	if _, err := query.Query(ctx, string(label), false); err != nil {
		if err := file.WriteFile(b.Path, originalContents); err != nil {
			return fmt.Errorf("failed to restore BUILD file in %v", err)
		}
		return fmt.Errorf("post-edit query failed, reverting %v", err)
	}
	return nil
}
