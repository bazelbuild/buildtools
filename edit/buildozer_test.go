/*
Copyright 2016 Google Inc. All Rights Reserved.
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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"testing"
)

func findBuildozer(root string) string {
	// The path for buildozer is arch specific, just search
	// in the data dir
	var ret string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() == "buildozer" && !info.IsDir() {
			ret = path
		}
		return nil
	})
	return ret
}

func getExitCode(err error) int {
	if e2, ok := err.(*exec.ExitError); ok {
		if status, ok := e2.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}

func TestPrintUsesDefaultSeparators(t *testing.T) {
	tmpDir, tmpErr := ioutil.TempDir("", "buildozer_test")
	if tmpErr != nil {
		panic(tmpErr)
	}
	defer os.RemoveAll(tmpDir)

	ioutil.WriteFile(
		path.Join(tmpDir, "BUILD"),
		[]byte(
			"some_rule("+
				"name=\"foo\", "+
				"list_attr=[\"bar\", \"baz\"], "+
				"attr=\"attr_value\")\n"+
				"some_rule(name=\"foo2\")"), 0644)

	expected := ("foo [bar baz] attr_value\n" +
		"foo2 (missing) (missing)\n")
	command := exec.Command(
		findBuildozer(os.Getenv("TEST_SRCDIR")),
		"print name list_attr attr",
		"//:*",
	)
	command.Dir = tmpDir

	out, err := command.Output()
	if err != nil && getExitCode(err) != 3 {
		t.Fatalf("Error running buildozer: %v", err)
	}

	outStr := string(out)
	if outStr != expected {
		t.Fatalf("Got %v, expected %v", outStr, expected)
	}
}

func TestPrintHonorsSeparators(t *testing.T) {
	tmpDir, tmpErr := ioutil.TempDir("", "buildozer_test")
	if tmpErr != nil {
		panic(tmpErr)
	}
	defer os.RemoveAll(tmpDir)

	ioutil.WriteFile(
		path.Join(tmpDir, "BUILD"),
		[]byte(
			"some_rule("+
				"name=\"foo\", "+
				"list_attr=[\"bar\", \"baz\"], "+
				"attr=\"attr_value\")\n"+
				"some_rule(name=\"foo2\")"), 0644)

	expected := ("foo\x01[bar,baz]\x01attr_value\x02" +
		"foo2\x01(missing)\x01(missing)\x02")
	command := exec.Command(
		findBuildozer(os.Getenv("TEST_SRCDIR")),
		"-print_list_separator",
		",",
		"-print_field_separator",
		"\x01",
		"-print_line_separator",
		"\x02",
		"print name list_attr attr",
		"//:*",
	)
	command.Dir = tmpDir

	out, err := command.Output()
	if err != nil && getExitCode(err) != 3 {
		t.Fatalf("Error running buildozer: %v", err)
	}

	outStr := string(out)
	if outStr != expected {
		t.Fatalf("Got %v, expected %v", outStr, expected)
	}
}
