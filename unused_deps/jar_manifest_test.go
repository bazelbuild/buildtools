/*
Copyright 2018 Google Inc. All Rights Reserved.
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
package main

import (
	"errors"
	"testing"
)

func TestManifestIndexNewline(t *testing.T) {
	foo := []struct {
		manifest string
		newline  int
		next     int
		err      error
	}{
		{"\n", 0, 1, nil},
		{"\r", 0, 1, nil},
		{"\n\r", 0, 1, nil},
		{"\r\n", 0, 2, nil},
		{"a\na", 1, 2, nil},
		{"a\ra", 1, 2, nil},
		{"a\nb\ra", 1, 2, nil},
		{"a\rb\na", 1, 2, nil},
		{"a\n\rb\ra", 1, 2, nil},
		{"a\r\nb\na", 1, 3, nil},
		{"", -1, -1, errors.New("no newline in ''")},
		{"aaa", -1, -1, errors.New("no newline in 'aaa'")},
	}

	errMismatch := func(e1, e2 error) bool {
		return (e1 != nil || e2 != nil) && (e1 == nil || e2 == nil || e1.Error() != e2.Error())
	}

	for i, tt := range foo {
		newline, next, err := manifestIndexNewline(tt.manifest)
		if newline != tt.newline || next != tt.next || errMismatch(err, tt.err) {
			t.Errorf("%d. manifestIndexNewline(%q) => (%d, %d, %q), want (%d, %d, %q)",
				i, tt.manifest, newline, next, err, tt.newline, tt.next, tt.err)
		}
	}
}
