/*
Copyright 2024 Google LLC

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

// Package adddep contains utilities for adding deps.
package adddep

import (
	"fmt"
	"regexp"

	"github.com/bazelbuild/buildtools/add_dep/bazel"
)

var suffixes = map[string]string{
	"java_proto_library":         "java_proto",
	"java_lite_proto_library":    "java_proto_lite",
	"java_mutable_proto_library": "java_proto_mutable",
	"kt_jvm_proto_library":       "kt_proto",
	"kt_jvm_lite_proto_library":  "kt_proto_lite",
}

var loads = map[string]string{
	"kt_jvm_lite_proto_library": "//third_party/protobuf/build_defs:kt_jvm_proto_library.bzl",
	"kt_jvm_proto_library":      "//third_party/protobuf/build_defs:kt_jvm_proto_library.bzl",
}

var protoRE = regexp.MustCompile("(?:_|^)(proto_v1|proto_v2|proto1|proto2|proto|protos|v1|v2)$")

// LangProtoLibraryName returns the proto wrapper rule to create for the given proto_library
// target and aspect.
func LangProtoLibraryName(target, aspect string) (string, error) {
	suffix, ok := suffixes[aspect]
	if !ok {
		return "", fmt.Errorf("unknown aspect name %s", aspect)
	}
	match := protoRE.FindStringIndex(target)
	if len(match) >= 2 {
		return target[:match[0]+1] + suffix + target[match[1]:], nil
	}
	return target + "_" + suffix, nil
}

// LangProtoLibraryLoad returns the file the given proto aspect rule kind needs to be loaded from, if any.
func LangProtoLibraryLoad(aspect string) (string, bool) {
	load, ok := loads[aspect]
	return load, ok
}

// LangProtoLibraryQuery returns the query to find existing proto aspect rules for the given proto.
func LangProtoLibraryQuery(aspect string, to bazel.Label) string {
	switch aspect {
	case "kt_jvm_lite_proto_library":
		fallthrough
	case "kt_jvm_proto_library":
		return fmt.Sprintf("attr(tags, %s, kind(kt_proto_library_helper, siblings(%s)))", aspect, to)
	default:
		return fmt.Sprintf("kind(%s, siblings(%s))", aspect, to)
	}
}
