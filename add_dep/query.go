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

// Package query contains a wrapper around bazel query.
package query

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bazelbuild/buildtools/add_dep/bazel"
	"github.com/bazelbuild/buildtools/add_dep/pipe"

	"github.com/golang/protobuf/proto"

	pb "github.com/bazelbuild/buildtools/build_proto"
)

// Rule is a rule proto and canonicalized bazel label returned by a query.
type Rule struct {
	pb    *pb.Rule
	Label bazel.Label
}

// HasDep returns true if the given label is listed in this rule's
// deps attribute.
func (r Rule) HasDep(dep bazel.Label) bool {
	for _, d := range r.attrStrings("deps") {
		pkg, _ := r.Label.Split()
		l, err := bazel.ParseRelativeLabel(pkg, d)
		if err == nil && l == dep {
			return true
		}
	}
	return false
}

func (r Rule) attrStrings(name string) []string {
	for _, attr := range r.pb.GetAttribute() {
		if attr.GetName() != name {
			continue
		}
		return attr.GetStringListValue()
	}
	return nil
}

// IsExport returns true if the given label is exported by this rule.
func (r Rule) IsExport(other bazel.Label) bool {
	pkg, _ := r.Label.Split()
	var exports []string
	if r.pb.GetRuleClass() == "alias" {
		for _, a := range r.pb.GetAttribute() {
			if a.GetName() == "actual" {
				exports = []string{a.GetStringValue()}
				break
			}
		}
	} else {
		exports = r.attrStrings("exports")
	}
	for _, export := range exports {
		l, err := bazel.ParseRelativeLabel(pkg, export)
		if err == nil && l == other {
			return true
		}
	}
	return false
}

// HasInput returns true if the given label is an input of this rule.
func (r Rule) HasInput(other bazel.Label) bool {
	for _, input := range r.pb.GetRuleInput() {
		label, err := bazel.ParseAbsoluteLabel(input)
		if err == nil && label == other {
			return true
		}
	}
	return false
}

// AvoidDep returns true if this rule has an 'avoid_dep' tag.
func (r Rule) AvoidDep() bool {
	for _, tag := range r.attrStrings("tags") {
		if tag == "avoid_dep" {
			return true
		}
	}
	return false
}

// RuleClass returns the rule class, e.g. java_library.
func (r Rule) RuleClass() string {
	return r.pb.GetRuleClass()
}

// Location returns the BUILD file path and line number of the rule.
func (r Rule) Location() (string, int, error) {
	fields := strings.Split(r.pb.GetLocation(), ":")
	if len(fields) != 3 {
		return "", -1, fmt.Errorf("expected exactly 3 `:`-separated fields in %s", r.pb.GetLocation())
	}
	f := fields[0]
	// Trim readonly prefixes from rabbit query results
	f = strings.TrimPrefix(f, "/workspace/READONLY/google3/")
	line, err := strconv.Atoi(fields[1])
	if err != nil {
		return "", -1, fmt.Errorf("bad line number %s: %v", fields[1], err)
	}
	return f, line, nil
}

// Inputs returns labels of inputs to the rule.
func (r Rule) Inputs() []bazel.Label {
	var result []bazel.Label
	for _, input := range r.pb.GetRuleInput() {
		if label, err := bazel.ParseAbsoluteLabel(input); err == nil {
			result = append(result, label)
		}
	}
	return result
}

func stderrPrinter(stderr io.Reader) {
	first := true
	// Most of the output is going to be 'loading package' messages,
	// which we emit on the same line to avoid filling up the scroll buffer.
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.Contains(text, "Loading:") {
			fmt.Fprintf(os.Stderr, "\033[2K\r\033[2m%s\033[0m", text)
			first = false
		}
	}
	if !first {
		fmt.Fprintln(os.Stderr)
	}
}

// Query performs a bazel query and returns the resulting Rules keyed by label.
func Query(ctx context.Context, query string, keepGoing bool) (map[bazel.Label]Rule, error) {
	var cmd = exec.Command(
		"bazel",
		// Allow the Blaze server to exit after 5 seconds if we're the reason for it starting up.
		// go/bum#startup_options
		"--max_idle_secs=5",
		"query",
		"--color=no",
		"--curses=no",
		// the query can still return useful results of the transitive closure is borked
		fmt.Sprintf("--keep_going=%v", keepGoing),
		"--notool_deps",
		"--implicit_deps",
		"--output=proto",
		"--tool_tag=add_depgo",
		query)
	var stderrbuf bytes.Buffer
	var out bytes.Buffer
	stderrPipe, stderrWriter := io.Pipe()
	delayed := pipe.New(stderrWriter)
	cmd.Stderr = io.MultiWriter(&stderrbuf, delayed)
	cmd.Stdout = &out
	startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	childCtx, cancel := context.WithCancel(ctx)
	timer := time.NewTimer(1 * time.Second)

	// Wait 1s and then start streaming output to stderr.
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		select {
		case <-timer.C:
			// The query was slow, so pacify the user with some helpful log output.
			go func() {
				fmt.Fprintf(os.Stderr, "\033[2mRunning bazel query: %s\033[0m\n", query)
				stderrPrinter(stderrPipe)
				elapsed := time.Since(startTime).Nanoseconds() / int64(time.Millisecond)
				fmt.Fprintf(os.Stderr, "\033[2K\r\033[2mFinished query in %dms\033[0m\n", elapsed)
				wg.Done()
			}()
			delayed.Start()
		case <-ctx.Done():
			wg.Done()
		}
	}(childCtx)

	err = cmd.Wait()
	stderrWriter.Close()
	cancel()
	wg.Wait()

	if err != nil {
		fmt.Fprint(os.Stderr, stderrbuf.String())
		exitCode := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exitCode = waitStatus.ExitStatus()
		}
		if exitCode == 3 && keepGoing {
			// exit code 3 indicates partial success (go/bazel-exit-codes)
		} else {
			return nil, fmt.Errorf("query failed: %v", err)
		}
	}
	qr := &pb.QueryResult{}
	if err := proto.Unmarshal(out.Bytes(), qr); err != nil {
		return nil, fmt.Errorf("unmarshalling result failed: %v", err)
	}
	rules := make(map[bazel.Label]Rule)
	for _, t := range qr.GetTarget() {
		if t.GetType() == pb.Target_RULE {
			label, err := bazel.ParseAbsoluteLabel(t.GetRule().GetName())
			if err != nil {
				return nil, fmt.Errorf("could not parse label %s: %v", t.GetRule().GetName(), err)
			}
			rules[label] = Rule{t.GetRule(), label}
		}
	}
	return rules, nil
}
