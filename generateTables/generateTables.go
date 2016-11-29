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
// generateTables is a tool that generates a go file from the Build language proto file.
// It generates a Go map to find the type of an attribute.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	blaze "github.com/bazelbuild/buildifier/build_proto"
	"github.com/golang/protobuf/proto"
)

// blazeBuildLanguage reads a proto file and returns a BuildLanguage object.
func blazeBuildLanguage(file string) (*blaze.BuildLanguage, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read %s: %s\n", file, err)
		return nil, err
	}

	lang := &blaze.BuildLanguage{}
	if err := proto.Unmarshal(data, lang); err != nil {
		return nil, err
	}
	return lang, nil
}

// generateTable returns a map that associate a type for each attribute name found in Blaze.
func generateTable(rules []*blaze.RuleDefinition) map[string]blaze.Attribute_Discriminator {
	types := make(map[string]blaze.Attribute_Discriminator)
	for _, r := range rules {
		for _, attr := range r.Attribute {
			types[*attr.Name] = *attr.Type
		}
	}

	// Because of inconsistencies in bazel, we need a few exceptions.
	types["resources"] = blaze.Attribute_LABEL_LIST
	types["out"] = blaze.Attribute_STRING
	types["outs"] = blaze.Attribute_STRING_LIST
	types["stamp"] = blaze.Attribute_TRISTATE
	types["strip"] = blaze.Attribute_BOOLEAN

	// Surprisingly, the name argument is missing.
	types["name"] = blaze.Attribute_STRING

	// package arguments are also not listed in the proto file
	types["default_hdrs_check"] = blaze.Attribute_STRING
	types["default_visibility"] = types["visibility"]
	types["default_copts"] = types["copts"]
	types["default_deprecation"] = types["deprecation"]
	types["default_testonly"] = types["testonly"]
	types["features"] = blaze.Attribute_STRING_LIST

	types["extra_srcs"] = types["srcs"]
	types["pytype_deps"] = types["deps"]

	return types
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Expected argument: proto file\n")
	}
	lang, err := blazeBuildLanguage(os.Args[1])
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	types := generateTable(lang.Rule)

	// sort the keys to get deterministic output
	keys := make([]string, 0, len(types))
	for i := range types {
		keys = append(keys, i)
	}
	sort.Strings(keys)

	// print
	fmt.Printf(`// Generated file, do not edit.
package lang

import blaze "github.com/bazelbuild/buildifier/build_proto"

var TypeOf = map[string]blaze.Attribute_Discriminator{
`)
	for _, attr := range keys {
		fmt.Printf("	\"%s\":	blaze.Attribute_%s,\n", attr, types[attr])
	}
	fmt.Printf("}\n")
}
