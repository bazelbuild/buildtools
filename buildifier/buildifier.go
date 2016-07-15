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

// Buildifier, a tool to parse and format BUILD files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"

	core "github.com/bazelbuild/buildifier/core"
)

const (
	successAppErrCode  int = iota // 0: success, everything went well
	syntaxAppErrCode              // 1: syntax errors in input
	usageAppErrCode               // 2: usage errors: invoked incorrectly
	internalAppErrCode            // 3: unexpected runtime errors: file I/O problems or internal bugs
)

var (
	// Undocumented; for debugging.
	showlog = flag.Bool("showlog", false, "show log in check mode")

	vflag = flag.Bool("v", false, "print verbose information on standard error")
	dflag = flag.Bool("d", false, "alias for -mode=diff")
	mode  = flag.String("mode", "", "formatting mode: check, diff, or fix (default fix)")
	path  = flag.String("path", "", "assume BUILD file has this path relative to the workspace directory")

	// Debug flags passed through to rewrite.go
	allowSort = stringList("allowsort", "additional sort contexts to treat as safe")
	disable   = stringList("buildifier_disable", "list of buildifier rewrites to disable")
)

func stringList(name, help string) func() []string {
	f := flag.String(name, "", help)
	return func() []string {
		return strings.Split(*f, ",")
	}
}

func exit(code int, format string, args ...interface{}) {
	if code != syntaxAppErrCode {
		format = "buildifier: " + format
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

func info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func usage() {
	exit(usageAppErrCode, `usage: buildifier [-d] [-v] [-mode=mode] [-path=path] [files...]

Buildifier applies a standard formatting to the named BUILD files.
The mode flag selects the processing: check, diff, or fix.
In check mode, buildifier prints a list of files that need reformatting.
In diff mode, buildifier shows the diffs that it would make.
In fix mode, buildifier updates the files that need reformatting and,
if the -v flag is given, prints their names to standard error.
The default mode is fix. -d is an alias for -mode=diff.

If no files are listed, buildifier reads a BUILD file from standard input. In
fix mode, it writes the reformatted BUILD file to standard output, even if no
changes are necessary.

Buildifier's reformatting depends in part on the path to the file relative
to the workspace directory. Normally buildifier deduces that path from the
file names given, but the path can be given explicitly with the -path
argument. This is especially useful when reformatting standard input,
or in scripts that reformat a temporary copy of a file.
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	// Pass down debug flags into build package
	core.DisableRewrites = disable()
	core.AllowSort = allowSort()

	if *dflag {
		if *mode != "" {
			exit(usageAppErrCode, "cannot specify both -d and -mode flags")
		}
		*mode = "diff"
	}

	// Check mode.
	switch *mode {
	default:
		exit(usageAppErrCode, "unrecognized mode %s; valid modes are check, diff, fix", *mode)
	case "":
		*mode = "fix"

	case "check", "diff", "fix":
		// ok
	}

	// If the path flag is set, must only be formatting a single file.
	// It doesn't make sense for multiple files to have the same path.
	if *path != "" && len(args) > 1 {
		exit(usageAppErrCode, "can only format one file when using -path flag")
	}

	var errors []error
	if len(args) == 0 || len(args) == 1 && args[0] == "-" {
		// Read from stdin, write to stdout.
		if *mode == "fix" {
			*mode = "pipe"
		}
		if *mode == "diff" {
			errors = diffFiles([]string{"stdin"})
		} else {
			errors = processFiles([]string{"stdin"})
		}
	} else {
		if *mode == "diff" {
			errors = diffFiles(args)
		} else {
			errors = processFiles(args)
		}
	}

	for _, e := range errors {
		if !core.IsSyntaxError(e) {
			// not a syntax error. Quit
			exit(internalAppErrCode, "internal error:", e)
		}

		info(e.Error())
	}

	if len(errors) > 0 {
		os.Exit(syntaxAppErrCode)
	}

	os.Exit(successAppErrCode)
}

func processFiles(files []string) []error {
	// Start nworker workers reading stripes of the input
	// argument list and sending the resulting data on
	// separate channels. file[k] is read by worker k%nworker
	// and delivered on ch[k%nworker].
	type result struct {
		file string
		err  error
	}

	if len(files) == 0 {
		// nothing to process
		return nil
	}

	var wg sync.WaitGroup

	in := make(chan string)
	ch := make(chan result, len(files))

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for filename := range in {
				err := processFile(filename)
				if err != nil {
					ch <- result{filename, err}
				}
			}
			wg.Done()
		}()
	}
	for _, fname := range files {
		in <- fname
	}
	close(in)
	wg.Wait()
	close(ch)

	var errors []error
	for res := range ch {
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
	}

	return errors
}

func read(filename string) (out []byte, err error) {
	if filename == "stdin" || filename == "-" {
		out, err = ioutil.ReadAll(os.Stdin)
	} else {
		out, err = ioutil.ReadFile(filename)
	}

	return
}

// processFile processes a single file containing data.
// It has been read from filename and should be written back if fixing.
func processFile(filename string) error {

	data, err := read(filename)

	f, err := core.Parse(filename, data)
	if err != nil {
		return err
	}

	if *path != "" {
		f.Path = *path
	}
	beforeRewrite := core.Format(f)
	rewriteInfo := core.Rewrite(f)
	ndata := core.Format(f)

	switch *mode {
	case "check":
		// check mode: print names of files that need formatting.
		if bytes.Equal(data, ndata) {
			return nil
		}
		// Print:
		//	name # list of what changed
		var (
			reformat string
			log      string
		)

		if !bytes.Equal(data, beforeRewrite) {
			reformat = " reformat"
		}

		if *showlog {
			set := make(map[string]struct{})

			for _, l := range rewriteInfo.Log {
				set[l] = struct{}{}
			}

			uniq := make([]string, 0, len(set))
			for k := range set {
				uniq = append(uniq, k)
			}

			log = " " + strings.Join(uniq, " ")
		}

		info("%s #%s %s%s\n", filename, reformat, rewriteInfo, log)
		return nil

	case "pipe":
		// pipe mode - reading from stdin, writing to stdout.
		// ("pipe" is not from the command line; it is set above in main.)
		os.Stdout.Write(ndata)
		return nil

	case "fix":
		// fix mode: update files in place as needed.
		if bytes.Equal(data, ndata) {
			return nil
		}

		err := ioutil.WriteFile(filename, ndata, 0666)
		if err != nil {
			return err
		}

		if *vflag {
			fmt.Fprintf(os.Stderr, "fixed %s\n", filename)
		}
	default:
		exit(usageAppErrCode, "mode not supported: "+*mode)
	}
	return nil
}
