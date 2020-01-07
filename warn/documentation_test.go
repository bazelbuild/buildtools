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
