/*
Copyright 2019 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package edit

import (
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
