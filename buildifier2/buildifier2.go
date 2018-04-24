// An alternative implementation of Buildifier, on top of the Skylark parser.
// https://github.com/google/skylark/

// This is experimental.

// If the experiment is successful, we might drop the AST defined in this
// package and use the AST from github.com/google/skylark/. This will give
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
	"github.com/google/skylark/syntax"
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
