/*
Copyright 2020 Google LLC

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

package warn

import (
	"github.com/bazelbuild/buildtools/tables"
	"testing"
)

func TestWarnLoadLocation(t *testing.T) {
	tables.RuleLoadLocation["s1"] = ":z.bzl"
	tables.RuleLoadLocation["s3"] = ":y.bzl"
	checkFindingsAndFix(t, "rule-load-location", `
load(":f.bzl", "s1", "s2")
load(":a.bzl", "s3")
`, `
load(":f.bzl", "s1", "s2")
load(":a.bzl", "s3")
`,
		[]string{
			":1: Rule \"s1\" must be loaded from :z.bzl.",
			":2: Rule \"s3\" must be loaded from :y.bzl.",
		},
		scopeEverywhere)
}
