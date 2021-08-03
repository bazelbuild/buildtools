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

package edit

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestEmptyCommandFileContainsNoCommands(t *testing.T) {
	reader := strings.NewReader("")
	commandsByBuildFile := make(map[string][]commandsForTarget)
	appendCommandsFromReader(NewOpts(), reader, commandsByBuildFile, []string{})
	t.Logf("Read commands:\n%s", prettyFormat(commandsByBuildFile))

	if len(commandsByBuildFile) != 0 {
		t.Error("No commands should be read")
	}
}

func TestCommandFileWithOneSetAttributeLineContainsOneCommand(t *testing.T) {
	commands := parseCommandFile("set srcs mytarget.go|//test-project:mytarget\n", t)
	if len(commands) != 1 {
		t.Error("Exactly one command should be read")
		return
	}
	if commands[0].target != "//test-project:mytarget" {
		t.Error("Read command should be for the correct target")
	}
	if !reflect.DeepEqual(commands[0].command, command{[]string{"set", "srcs", "mytarget.go"}}) {
		t.Error("Read command should contain the correct command tokens")
	}
}

func TestCommandFileWithTwoSetAttributesInOneLineContainsTwoCommands(t *testing.T) {
	commands := parseCommandFile("set srcs mytarget.go|set deps //other-project:othertarget|//test-project:mytarget\n", t)
	if len(commands) != 2 {
		t.Error("Exactly two commands should be read")
		return
	}
	if commands[0].target != "//test-project:mytarget" {
		t.Error("First read command should be for the correct target")
	}
	if !reflect.DeepEqual(commands[0].command, command{[]string{"set", "srcs", "mytarget.go"}}) {
		t.Error("First read command should contain the correct command tokens")
	}
	if commands[1].target != "//test-project:mytarget" {
		t.Error("Second read command should be for the correct target")
	}
	if !reflect.DeepEqual(commands[1].command, command{[]string{"set", "deps", "//other-project:othertarget"}}) {
		t.Error("Second read command should contain the correct command tokens")
	}
}

func TestCommandFileWithTwoSetAttributesInSeparateLinesContainsTwoCommands(t *testing.T) {
	commands := parseCommandFile("set srcs mytarget.go|//test-project:mytarget\nset deps //other-project:othertarget|//test-project:mytarget\n", t)
	if len(commands) != 2 {
		t.Error("Exactly two commands should be read")
		return
	}
	if commands[0].target != "//test-project:mytarget" {
		t.Error("First read command should be for the correct target")
	}
	if !reflect.DeepEqual(commands[0].command, command{[]string{"set", "srcs", "mytarget.go"}}) {
		t.Error("First read command should contain the correct command tokens")
	}
	if commands[1].target != "//test-project:mytarget" {
		t.Error("Second read command should be for the correct target")
	}
	if !reflect.DeepEqual(commands[1].command, command{[]string{"set", "deps", "//other-project:othertarget"}}) {
		t.Error("Second read command should contain the correct command tokens")
	}
}

func TestCommandFileWithoutTrailingNewlineContainsCommand(t *testing.T) {
	commands := parseCommandFile("set srcs mytarget.go|//test-project:mytarget", t)
	if len(commands) != 1 {
		t.Error("Exactly one command should be read")
		return
	}
	if commands[0].target != "//test-project:mytarget" {
		t.Error("Read command should be for the correct target")
	}
}

func TestBlankLinesInCommandFileAreIgnored(t *testing.T) {
	commands := parseCommandFile("set srcs mytarget.go|//test-project:mytarget\n\n\n\n\nset srcs othertarget.go|//test-project:othertarget\n", t)
	if len(commands) != 2 {
		t.Error("Exactly two commands should be read")
	}
}

func TestLongLineInCommandFileParsesAsOneCommand(t *testing.T) {
	srcsLength := 10000

	expectedCommandTokens := make([]string, srcsLength+2)
	expectedCommandTokens[0] = "set"
	expectedCommandTokens[1] = "srcs"
	srcs := make([]string, srcsLength)
	for i := 0; i < srcsLength; i++ {
		src := "source_" + strconv.Itoa(i) + ".go"
		srcs[i] = src
		expectedCommandTokens[i+2] = src
	}

	commands := parseCommandFile("set srcs "+strings.Join(srcs, " ")+"|//test-project:mytarget\n", t)
	if len(commands) != 1 {
		t.Error("Exactly one command should be read")
		return
	}
	if !reflect.DeepEqual(commands[0].command, command{expectedCommandTokens}) {
		t.Errorf("First read command should contain the correct command tokens")
	}
}

func parseCommandFile(fileContent string, t *testing.T) []parsedCommand {
	reader := strings.NewReader(fileContent)
	commandsByBuildFile := make(map[string][]commandsForTarget)
	appendCommandsFromReader(NewOpts(), reader, commandsByBuildFile, []string{})
	t.Logf("Read commands:\n%s", prettyFormat(commandsByBuildFile))
	return extractCommands(commandsByBuildFile)
}

type parsedCommand struct {
	command   command
	target    string
	buildFile string
}

func extractCommands(commandsByBuildFile map[string][]commandsForTarget) []parsedCommand {
	out := make([]parsedCommand, 0)
	for buildFile, targets := range commandsByBuildFile {
		for _, target := range targets {
			for _, command := range target.commands {
				out = append(out, parsedCommand{command, target.target, buildFile})
			}
		}
	}
	return out
}

func prettyFormat(commandsByBuildFile map[string][]commandsForTarget) string {
	out := ""
	for buildFile, targets := range commandsByBuildFile {
		out += buildFile + "\n"
		for _, target := range targets {
			out += "  target: " + target.target + "\n"
			for _, command := range target.commands {
				out += "    -"
				for _, token := range command.tokens {
					out += " " + token
				}
				out += "\n"
			}
		}
	}
	return out
}
