/*
Copyright 2019 Google LLC

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

package edit

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/build"
)

var removeCommentTests = []struct {
	args      []string
	buildFile string
	expected  string
}{
	{[]string{},
		`# comment
	foo(
		name = "foo",
	)`,
		`foo(
    name = "foo",
)`,
	},
	{[]string{
		"name",
	},
		`foo(
		# comment
		name = "foo",
	)`,
		`foo(
    name = "foo",
)`,
	},
	{[]string{
		"name",
	},
		`foo(
		name = "foo"  # comment,
	)`,
		`foo(
    name = "foo",
)`,
	},
	{[]string{
		"deps", "bar",
	},
		`foo(
		name = "foo",
		deps = [
				# comment
				"bar",
				"baz",
		],
	)`,
		`foo(
    name = "foo",
    deps = [
        "bar",
        "baz",
    ],
)`,
	},
	{[]string{
		"deps", "bar",
	},
		`foo(
		name = "foo",
		deps = [
				"bar",  # comment
				"baz",
		],
	)`,
		`foo(
    name = "foo",
    deps = [
        "bar",
        "baz",
    ],
)`,
	},
}

func TestCmdRemoveComment(t *testing.T) {
	for i, tt := range removeCommentTests {
		bld, err := build.Parse("BUILD", []byte(tt.buildFile))
		if err != nil {
			t.Error(err)
			continue
		}
		rl := bld.Rules("foo")[0]
		env := CmdEnvironment{
			File: bld,
			Rule: rl,
			Args: tt.args,
		}
		bld, _ = cmdRemoveComment(NewOpts(), env)
		got := strings.TrimSpace(string(build.Format(bld)))
		if got != tt.expected {
			t.Errorf("cmdRemoveComment(%d):\ngot:\n%s\nexpected:\n%s", i, got, tt.expected)
		}
	}
}

type targetExpressionToBuildFilesTestCase struct {
	rootDir, target string
	buildFiles      []string
}

func setupTestTmpWorkspace(t *testing.T, buildFileName string) (tmp string) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}

	// On MacOS "/tmp" is a symlink to "/private/tmp". Resolve it to make the testing easier
	tmp, err = filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(tmp, "a", "b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "a", "c"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create additional directories that will be ignored
	if err := os.MkdirAll(filepath.Join(tmp, "ignored"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "a", "ignored"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "WORKSPACE"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", "b", buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", "c", buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}
	// Create BUILD files in ignored directories to verify they're skipped
	if err := os.WriteFile(filepath.Join(tmp, "ignored", buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", "ignored", buildFileName), nil, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .bazelignore file with paths to ignore
	bazelignoreContent := []byte("# Ignore these directories\nignored\na/ignored\n")
	if err := os.WriteFile(filepath.Join(tmp, ".bazelignore"), bazelignoreContent, 0644); err != nil {
		t.Fatal(err)
	}

	return
}

func runTestTargetExpressionToBuildFiles(t *testing.T, buildFileName string) {
	tmp := setupTestTmpWorkspace(t, buildFileName)
	defer os.RemoveAll(tmp)

	for _, tc := range []targetExpressionToBuildFilesTestCase{
		{tmp, "//", []string{filepath.Join(tmp, buildFileName)}},
		{tmp, "//:foo", []string{filepath.Join(tmp, buildFileName)}},
		{tmp, "//a", []string{filepath.Join(tmp, "a", buildFileName)}},
		{tmp, "//a:foo", []string{filepath.Join(tmp, "a", buildFileName)}},
		{tmp, "//a/b", []string{filepath.Join(tmp, "a", "b", buildFileName)}},
		{tmp, "//a/b:foo", []string{filepath.Join(tmp, "a", "b", buildFileName)}},
		{tmp, "//...", []string{filepath.Join(tmp, buildFileName), filepath.Join(tmp, "a", buildFileName), filepath.Join(tmp, "a", "b", buildFileName), filepath.Join(tmp, "a", "c", buildFileName)}},
		{tmp, "//a/...", []string{filepath.Join(tmp, "a", buildFileName), filepath.Join(tmp, "a", "b", buildFileName), filepath.Join(tmp, "a", "c", buildFileName)}},
		{tmp, "//a/b/...", []string{filepath.Join(tmp, "a", "b", buildFileName)}},
		{tmp, "//a/c/...", []string{filepath.Join(tmp, "a", "c", buildFileName)}},
		{tmp, "//a/c/...:foo", []string{filepath.Join(tmp, "a", "c", buildFileName)}},
		{"", "...:foo", []string{filepath.Join(tmp, buildFileName), filepath.Join(tmp, "a", buildFileName), filepath.Join(tmp, "a", "b", buildFileName), filepath.Join(tmp, "a", "c", buildFileName)}},
	} {
		if tc.rootDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			// buildozer should be able to find the WORKSPACE file in the current wd
			if err := os.Chdir(tmp); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(cwd)
		}

		buildFiles := targetExpressionToBuildFiles(tc.rootDir, tc.target)
		expectedBuildFilesMap := make(map[string]bool)
		buildFilesMap := make(map[string]bool)
		for _, buildFile := range buildFiles {
			buildFilesMap[buildFile] = true
		}
		for _, buildFile := range tc.buildFiles {
			expectedBuildFilesMap[buildFile] = true
		}
		if !reflect.DeepEqual(expectedBuildFilesMap, buildFilesMap) {
			t.Errorf("TargetExpressionToBuildFiles(%q, %q) = %q want %q", tc.rootDir, tc.target, buildFiles, tc.buildFiles)
		}
	}
}

func TestTargetExpressionToBuildFiles(t *testing.T) {
	for _, buildFileName := range BuildFileNames {
		runTestTargetExpressionToBuildFiles(t, buildFileName)
	}
}

func runTestAppendCommands(t *testing.T, buildFileName string) {
	tmp := setupTestTmpWorkspace(t, buildFileName)
	defer os.RemoveAll(tmp)

	for _, tc := range []targetExpressionToBuildFilesTestCase{
		{tmp, ".:__pkg__", []string{"./" + buildFileName}},
		{tmp, "a" + ":__pkg__", []string{"a/" + buildFileName}},
		{"", "a" + ":__pkg__", []string{"a/" + buildFileName}},
	} {
		if tc.rootDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			// buildozer should be able to find the WORKSPACE file in the current wd
			if err := os.Chdir(tmp); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(cwd)
		}

		commandsByFile := make(map[string][]commandsForTarget)
		opts := NewOpts()
		opts.RootDir = tc.rootDir
		appendCommands(opts, commandsByFile, tc.buildFiles)
		if len(commandsByFile) != 1 {
			t.Errorf("Expect one target after appendCommands")
		}
		for _, value := range commandsByFile {
			if value[0].target != tc.target {
				t.Errorf("appendCommands for buildfile %s yielded target %s, expected %s", tc.buildFiles, value[0].target, tc.target)
			}
		}
	}
}

func TestAppendCommands(t *testing.T) {
	for _, buildFileName := range BuildFileNames {
		runTestAppendCommands(t, buildFileName)
	}
}

var dictListAddTests = []struct {
	args      []string
	buildFile string
	expected  string
}{
	{[]string{
		"attr", "key1", "value1",
	},
		`foo(
		name = "foo",
	)`,
		`foo(
    name = "foo",
    attr = {"key1": ["value1"]},
)`,
	},
	{[]string{
		"attr", "key1", "value2",
	},
		`foo(
		name = "foo",
		attr = {"key1": ["value1"]},
	)`,
		`foo(
    name = "foo",
    attr = {"key1": [
        "value1",
        "value2",
    ]},
)`,
	},
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
	)`,
		`foo(
    name = "foo",
    attr = {"key1": [
        "value1",
        "value2",
    ]},
)`,
	},
	{[]string{
		"attr", "key2", "value2",
	},
		`foo(
		name = "foo",
		attr = {"key1": ["value1"]},
	)`,
		`foo(
    name = "foo",
    attr = {
        "key1": ["value1"],
        "key2": ["value2"],
    },
)`,
	},
	{[]string{
		"attr", "key1", "value1",
	},
		`foo(
		name = "foo",
		attr = {"key1": ["value1"]},
	)`,
		`foo(
    name = "foo",
    attr = {"key1": ["value1"]},
)`,
	},
}

func TestCmdDictListAdd(t *testing.T) {
	for i, tt := range dictListAddTests {
		bld, err := build.Parse("BUILD", []byte(tt.buildFile))
		if err != nil {
			t.Error(err)
			continue
		}
		rl := bld.Rules("foo")[0]
		env := CmdEnvironment{
			File: bld,
			Rule: rl,
			Args: tt.args,
		}
		bld, _ = cmdDictListAdd(NewOpts(), env)
		got := strings.TrimSpace(string(build.Format(bld)))
		if got != tt.expected {
			t.Errorf("cmdDictListAdd(%d):\ngot:\n%s\nexpected:\n%s", i, got, tt.expected)
		}
	}
}

var dictReplaceIfEqualTests = []struct {
	args      []string
	buildFile string
	expected  string
}{
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
		attr = {"key1": "value1"},
	)`,
		`foo(
    name = "foo",
    attr = {"key1": "value2"},
)`,
	},
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
		attr = {"key1": "x"},
	)`,
		`foo(
    name = "foo",
    attr = {"key1": "x"},
)`,
	},
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
		attr = {
      "key1": ["value1"],
			"key2": ["value2"],
	},
	)`,
		`foo(
    name = "foo",
    attr = {
			"key1": ["value1"],
			"key2": ["value2"],
    },
)`,
	},
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
		attr = {
			"key1": "value1",
			"key2": "value2",
	},
	)`,
		`foo(
		name = "foo",
		attr = {
			"key1": "value2",
			"key2": "value2",
},
)`,
	},
	{[]string{
		"attr", "key1", "value1", "value2",
	},
		`foo(
		name = "foo",
		attr = {
			"key1": "value1",
			"key2": "x",
	},
	)`,
		`foo(
    name = "foo",
    attr = {
			"key1": "value2",
			"key2": "x",
		},
)`,
	},
}

func TestCmdDictReplaceIfEqual(t *testing.T) {
	for i, tt := range dictReplaceIfEqualTests {
		bld, err := build.Parse("BUILD", []byte(tt.buildFile))
		if err != nil {
			t.Error(err)
			continue
		}
		expectedBld, err := build.Parse("BUILD", []byte(tt.expected))
		if err != nil {
			t.Error(err)
			continue
		}
		rl := bld.Rules("foo")[0]
		env := CmdEnvironment{
			File: bld,
			Rule: rl,
			Args: tt.args,
		}
		bld, err = cmdDictReplaceIfEqual(NewOpts(), env)
		if err != nil {
			t.Errorf("cmdDictReplaceIfEqual(%d):\ngot error:\n%s", i, err)
		}
		got := strings.TrimSpace(string(build.Format(bld)))
		expected := strings.TrimSpace(string(build.Format(expectedBld)))
		if got != expected {
			t.Errorf("cmdDictReplaceIfEqual(%d):\ngot:\n%s\nexpected:\n%s", i, got, tt.expected)
		}
	}
}

var substituteLoadsTests = []struct {
	args      []string
	buildFile string
	expected  string
}{
	{[]string{
		"^(.*)$", "${1}",
	},
		`load("//foo:foo.bzl", "foo")`,
		`load("//foo:foo.bzl", "foo")`,
	},
	{[]string{
		"^@rules_foo//foo:defs.bzl$", "//build/rules/foo:defs.bzl",
	},
		`load("@rules_bar//bar:defs.bzl", "bar")`,
		`load("@rules_bar//bar:defs.bzl", "bar")`,
	},
	{[]string{
		"^@rules_foo//foo:defs.bzl$", "//build/rules/foo:defs.bzl",
	},
		`load("@rules_foo//foo:defs.bzl", "foo", "foo2")
load("@rules_bar//bar:defs.bzl", "bar")`,
		`load("@rules_bar//bar:defs.bzl", "bar")
load("//build/rules/foo:defs.bzl", "foo", "foo2")`,
	},
	{[]string{
		":foo.bzl$", ":defs.bzl",
	},
		`load("//foo:foo.bzl", "foo")`,
		`load("//foo:defs.bzl", "foo")`,
	},
	{[]string{
		// Keep in sync with the example in `//buildozer:README.md`.
		"^@([^/]*)//([^:].*)$", "//third_party/build_defs/${1}/${2}",
	},
		`load("@rules_foo//foo:defs.bzl", "foo", "foo2")
load("@rules_bar//bar:defs.bzl", "bar")
load("@rules_bar//:defs.bzl", legacy_bar = "bar")`,
		`load("@rules_bar//:defs.bzl", legacy_bar = "bar")
load("//third_party/build_defs/rules_bar/bar:defs.bzl", "bar")
load("//third_party/build_defs/rules_foo/foo:defs.bzl", "foo", "foo2")`,
	},
}

func TestCmdSubstituteLoad(t *testing.T) {
	for i, tt := range substituteLoadsTests {
		bld, err := build.Parse("BUILD", []byte(tt.buildFile))
		if err != nil {
			t.Error(err)
			continue
		}
		env := CmdEnvironment{
			File: bld,
			Args: tt.args,
		}
		bld, _ = cmdSubstituteLoad(NewOpts(), env)
		got := strings.TrimSpace(string(build.Format(bld)))
		if got != tt.expected {
			t.Errorf("cmdSubstituteLoad(%d):\ngot:\n%s\nexpected:\n%s", i, got, tt.expected)
		}
	}
}

func TestCmdSubstitute(t *testing.T) {
	for i, tc := range []struct {
		name      string
		args      []string
		buildFile string
		expected  string
	}{
		{
			name:      "empty_rule",
			args:      []string{"*", "^$", "x"},
			buildFile: `cc_library()`,
			expected:  `cc_library()`,
		},
		{
			name:      "known_attr",
			args:      []string{"*", "^//(.*)$", "//foo/${1}"},
			buildFile: `cc_library(deps = ["//bar/baz:quux"])`,
			expected:  `cc_library(deps = ["//foo/bar/baz:quux"])`,
		},
		{
			name:      "custom_attr",
			args:      []string{"*", "^//(.*)$", "//foo/${1}"},
			buildFile: `cc_library(my_custom_attr = "//bar/baz:quux")`,
			expected:  `cc_library(my_custom_attr = "//foo/bar/baz:quux")`,
		},
		{
			name:      "specific_rule",
			args:      []string{"deps", "^//(.*)$", "//foo/${1}"},
			buildFile: `cc_library(deps = ["//bar"], fancy_deps = ["//bar/baz:quux"])`,
			expected: `cc_library(
    fancy_deps = ["//bar/baz:quux"],
    deps = ["//foo/bar"],
)`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bld, err := build.Parse("BUILD", []byte(tc.buildFile))
			if err != nil {
				t.Error(err)
				return
			}
			env := CmdEnvironment{
				File: bld,
				Args: tc.args,
				Rule: bld.RuleAt(1),
			}
			bld, _ = cmdSubstitute(NewOpts(), env)
			got := strings.TrimSpace(string(build.Format(bld)))
			if got != tc.expected {
				t.Errorf("cmdSubstitute(%d):\ngot:\n%s\nexpected:\n%s", i, got, tc.expected)
			}
		})
	}
}

func TestCmdDictAddSet_missingColon(t *testing.T) {
	for _, tc := range []struct {
		name string
		fun  func(*Options, CmdEnvironment) (*build.File, error)
	}{
		{"dict_add", cmdDictAdd},
		{"dict_set", cmdDictSet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bld, err := build.Parse("BUILD", []byte("rule()"))
			if err != nil {
				t.Fatal(err)
			}
			env := CmdEnvironment{
				File: bld,
				Rule: bld.RuleAt(1),
				Args: []string{"attr", "invalid"},
			}
			_, err = tc.fun(NewOpts(), env)
			if err == nil {
				t.Error("succeeded, want error")
			}
		})
	}
}

func TestCmdSetSelect(t *testing.T) {
	for i, tc := range []struct {
		name      string
		args      []string
		buildFile string
		expected  string
	}{
		{
			name: "select_statement_doesn't_exist",
			args: []string{
				"args",                                   /* attr */
				":use_ci_timeouts", "-test.timeout=123s", /* key, value */
				":use_ci_timeouts", "-test.anotherFlag=flagValue", /* key, value */
				"//conditions:default", "-test.timeout=789s", /* key, value */
			},
			buildFile: `foo(
			name = "foo",
)`,
			expected: `foo(
    name = "foo",
    args = select({
        ":use_ci_timeouts": [
            "-test.timeout=123s",
            "-test.anotherFlag=flagValue",
        ],
        "//conditions:default": ["-test.timeout=789s"],
    }),
)`},
		{
			name: "select_statement_exists",
			args: []string{
				"args",                                   /* attr */
				":use_ci_timeouts", "-test.timeout=543s", /* key, value */
				"//conditions:default", "-test.timeout=876s", /* key, value */
			},
			buildFile: `foo(
    name = "foo",
    args = select({
        ":use_ci_timeouts": [
            "-test.timeout=123s",
            "-test.anotherFlag=flagValue",
        ],
        "//conditions:default": ["-test.timeout=789s"],
    }),
)`,
			expected: `foo(
    name = "foo",
    args = select({
        ":use_ci_timeouts": ["-test.timeout=543s"],
        "//conditions:default": ["-test.timeout=876s"],
    }),
)`},
		{
			name: "attr_exists_but_not_select",
			args: []string{
				"args",                                   /* attr */
				":use_ci_timeouts", "-test.timeout=543s", /* key, value */
				"//conditions:default", "-test.timeout=876s", /* key, value */
			},
			buildFile: `foo(
    name = "foo",
    args = ["-test.timeout=123s"],
)`,
			expected: `foo(
    name = "foo",
    args = select({
        ":use_ci_timeouts": ["-test.timeout=543s"],
        "//conditions:default": ["-test.timeout=876s"],
    }),
)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bld, err := build.Parse("BUILD", []byte(tc.buildFile))
			if err != nil {
				t.Error(err)
			}
			rl := bld.Rules("foo")[0]
			env := CmdEnvironment{
				File: bld,
				Rule: rl,
				Args: tc.args,
			}
			bld, _ = cmdSetSelect(NewOpts(), env)
			got := strings.TrimSpace(string(build.Format(bld)))
			if got != tc.expected {
				t.Errorf("cmdSetSelect(%d):\ngot:\n%s\nexpected:\n%s", i, got, tc.expected)
			}
		})
	}
}

func TestGetIgnoredPrefixes(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	tests := []struct {
		name          string
		bazelignore   string
		expected      []string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid paths",
			bazelignore: `# Ignore these directories
ignored
a/ignored
b/c/d`,
			expected: []string{"ignored", "a/ignored", "b/c/d"},
		},
		{
			name:        "empty file",
			bazelignore: ` `,
			expected:    []string{},
		},
		{
			name: "only comments",
			bazelignore: `# This is a comment
# Another comment`,
			expected: []string{},
		},
		{
			name: "empty lines",
			bazelignore: `ignored

a/ignored

# comment
b/c/d`,
			expected: []string{"ignored", "a/ignored", "b/c/d"},
		},
		{
			name: "absolute paths",
			bazelignore: `/absolute/path
ignored
/another/absolute/path`,
			expected: []string{"ignored"},
		},
		{
			name:        "no file",
			bazelignore: "",
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new temporary directory for each test case
			testDir, err := os.MkdirTemp(tmp, "")
			if err != nil {
				t.Fatal(err)
			}

			if tt.bazelignore != "" {
				if err := os.WriteFile(filepath.Join(testDir, ".bazelignore"), []byte(tt.bazelignore), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got := getIgnoredPrefixes(testDir)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("getIgnoredPrefixes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
