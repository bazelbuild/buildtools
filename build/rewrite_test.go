/*
Copyright 2022 Google LLC

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

package build

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var workingDir string = path.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), "build")
var originalFilePath = workingDir + "/rewrite_test_files/original.star"
var formattedFilePath = workingDir + "/rewrite_test_files/original_formatted.star"

var originalBytes, _ = ioutil.ReadFile(originalFilePath)
var originalFile, _ = ParseDefault(originalFilePath, originalBytes)

var formattedBytes, _ = ioutil.ReadFile(formattedFilePath)
var formattedFile, _ = ParseDefault(formattedFilePath, formattedBytes)

var name map[string]int = map[string]int{"name": -99}
var rewriteSet []string = []string{"callsort"}

var rewriter = Rewriter{
	RewriteSet:   rewriteSet,
	NamePriority: name,
}

func TestRewriterRegular(t *testing.T) {
	// Load the original file
	var modifyBytes, _ = ioutil.ReadFile(originalFilePath)
	var modifiedFile, _ = ParseDefault(originalFilePath, modifyBytes)

	// Perform rewrite on loaded file, rewrite should do nothing here
	Rewrite(modifiedFile)

	// Initialize printers to obtain bytes later for different types of printers
	// We will check bytes from printers because that is our source of truth before writing to new file
	formattedPrinter := &printer{fileType: formattedFile.Type}
	originalPrinter := &printer{fileType: originalFile.Type}
	modifiedPrinter := &printer{fileType: modifiedFile.Type}

	formattedPrinter.file(formattedFile)
	originalPrinter.file(originalFile)
	modifiedPrinter.file(modifiedFile)

	// Assert that bytes to be writter are same as original bytes and different from formatted
	if !bytes.Equal(originalPrinter.Bytes(), modifiedPrinter.Bytes()) {
		t.Error("Original Printer should equal Modified Printer")
	}
	if bytes.Equal(formattedPrinter.Bytes(), modifiedPrinter.Bytes()) {
		t.Error("Formmated Printer should not equal Modified Printer")
	}
}

func TestRewriterWithRewriter(t *testing.T) {
	// Load the original file
	var modifyBytes, _ = ioutil.ReadFile(originalFilePath)
	var modifiedFile, _ = ParseDefault(originalFilePath, modifyBytes)

	// Perform rewrite with rewriter on loaded file, should proceed to reorder
	rewriter.Rewrite(modifiedFile)

	// Initialize printers to obtain bytes later for different types of printers
	// We will check bytes from printers because that is our source of truth before writing to new file
	formattedPrinter := &printer{fileType: formattedFile.Type}
	originalPrinter := &printer{fileType: originalFile.Type}
	modifiedPrinter := &printer{fileType: modifiedFile.Type}

	formattedPrinter.file(formattedFile)
	originalPrinter.file(originalFile)
	modifiedPrinter.file(modifiedFile)

	// Assert that bytes to be written is same as formmatted bytes and different from original
	if !bytes.Equal(formattedPrinter.Bytes(), modifiedPrinter.Bytes()) {
		t.Error("Formatted Printer should equal Modified Printer")
	}
	if bytes.Equal(originalPrinter.Bytes(), modifiedPrinter.Bytes()) {
		t.Error("Original Printer should not equal Modified Printer")
	}
}
