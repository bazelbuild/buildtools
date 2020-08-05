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

// An alternative implementation of Buildifier, on top of the Skylark parser.
// https://go.starlark.net/

// This is experimental.

// If the experiment is successful, we might drop the AST defined in this
// package and use the AST from go.starlark.net/syntax. This will give
// us a much more precise AST and will allow us to share code with the
// Skylark interpreter. The end goal is to build a number of tools able
// to parse, analyze, format, refactor, evaluate Skylark code.

// Package main implements a buildifier on top of 'Skylark in Go'.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/convertast"
	"go.starlark.net/syntax"
)

func main() {
	flag.Parse()

	switch len(flag.Args()) {
	case 0:
		log.Fatal("Argument missing")
	case 1:
		filename := flag.Args()[0]
		ast, err := syntax.Parse(filename, nil, syntax.RetainComments)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		newAst := convertast.ConvFile(ast)
		fmt.Print(build.FormatString(newAst))

	default:
		log.Fatal("want at most one Skylark file name")
	}
}
