/*
Copyright 2026 Google LLC

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
	"slices"
	"strings"
)

// checkSelfOrInheritedComment checks comments relate to the expression, and returns true if
// the provided predicate matches any of the comments.
func checkSelfOrInheritedComment(expr Expr, predicate func(Comment) bool) bool {
	for _, comment := range slices.Concat(
		expr.Comment().Before,
		expr.Comment().After,
		expr.Comment().Suffix,
		expr.Comment().Inherited) {
		if predicate(comment) {
			return true
		}
	}
	return false
}

// HasCommentContaining does a case insensitive matching to see if an expression,
// or its parent expressions have a comment containing the provided prefix.
func HasCommentContaining(expr Expr, prefix string) bool {
	return checkSelfOrInheritedComment(expr, func(comment Comment) bool {
		trimmedComment := strings.Trim(comment.Token, " \t\n#")
		return strings.Contains(strings.ToLower(trimmedComment), strings.ToLower(prefix))
	})
}
