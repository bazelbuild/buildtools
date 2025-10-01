/*
Copyright 2017 Google LLC

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

package tables

import (
	"os"
	"reflect"
	"testing"
)

func TestParseJSONDefinitions(t *testing.T) {
	testdata := os.Getenv("TEST_SRCDIR") + "/" + os.Getenv("TEST_WORKSPACE") + "/tables/testdata"
	definitions, err := ParseJSONDefinitions(testdata + "/simple_tables.json")
	if err != nil {
		t.Error(err)
	}

	expected := Definitions{
		IsLabelArg:                 map[string]bool{"srcs": true},
		LabelDenylist:              map[string]bool{},
		IsSortableListArg:          map[string]bool{"srcs": true, "visibility": true},
		SortableDenylist:           map[string]bool{"genrule.srcs": true},
		SortableAllowlist:          map[string]bool{},
		NamePriority:               map[string]int{"name": -1},
		CompactConstantDefinitions: true,
		StripLabelLeadingSlashes:   true,
	}
	if !reflect.DeepEqual(expected, definitions) {
		t.Errorf("ParseJSONDefinitions(simple_tables.json) = %v; want %v", definitions, expected)
	}
}
