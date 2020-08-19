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
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/bazelbuild/buildtools/testutils"
)

func TestWarningsDocumentation(t *testing.T) {
	files, chdir := testutils.FindTests(t, "", "WARNINGS.md")
	defer chdir()

	if len(files) != 1 {
		t.Fatalf("Expected to find exactly one WARNINGS.md file, got %d instead", len(files))
	}
	data, err := ioutil.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
		return
	}

	contents := string(data)
	for _, warning := range AllWarnings {
		link := fmt.Sprintf("  * [`%s`](#%s)", warning, warning)
		if !strings.Contains(contents, link) {
			t.Errorf("No link (%q) found for the warning %q in WARNINGS.md, is it documented?", link, warning)
		}

		anchor := fmt.Sprintf(`<a name="%s"></a>`, warning)
		if !strings.Contains(contents, anchor) {
			t.Errorf("No anchor (%q) found for the warning %q in WARNINGS.md, is it documented?", anchor, warning)
		}
	}
}
