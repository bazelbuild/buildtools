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

// Package testutils provides some useful helpers for buildozer/buildifer tests.
package testutils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

// Diff returns the output of running diff on b1 and b2.
func Diff(b1, b2 []byte) ([]byte, error) {
	f1, err := ioutil.TempFile("", "testdiff")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "testdiff")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err := exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return data, err
}

// Tdiff logs the Diff output to t.Error.
func Tdiff(t *testing.T, a, b []byte) {
	data, err := Diff(a, b)
	if err != nil {
		t.Error(err)
		return
	}
	t.Error(string(data))
}
