/*
Copyright 2018 Bazel contributors. All Rights Reserved.

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

package build

import (
	"testing"
)

func TestIsBuildFilename(t *testing.T) {
	cases := map[string]bool{
		"BUILD":           true,
		"build":           true,
		"bUIld":           true,
		"BUILD.bazel":     true,
		"build.bzl":       false,
		"build.sky":       false,
		"WORKSPACE":       true,
		"external.BUILD":  true,
		"BUILD.external":  true,
		"aBUILD":          false,
		"thing.sky":       false,
		"my.WORKSPACE":    true,
		"thing.bzl":       false,
		"workspace.bazel": true,
	}
	for name, isBuild := range cases {
		res := isBuildFilename(name)
		if res != isBuild {
			t.Errorf("isBuildFilename(%q) should be %v but was %v", name, isBuild, res)
		}
	}
}
