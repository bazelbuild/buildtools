/*
Copyright 2021 Google LLC

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

// GetParamName extracts the param name from an item of function params.
func GetParamName(param Expr) (name, op string) {
	ident, op := GetParamIdent(param)
	if ident == nil {
		return "", ""
	}
	return ident.Name, op
}

// GetParamIdent extracts the param identifier from an item of function params.
func GetParamIdent(param Expr) (ident *Ident, op string) {
	switch param := param.(type) {
	case *Ident:
		return param, ""
	case *TypedIdent:
		return param.Ident, ""
	case *AssignExpr:
		// keyword parameter
		return GetParamIdent(param.LHS)
	case *UnaryExpr:
		// *args, **kwargs, or *
		if param.X == nil {
			// An asterisk separating position and keyword-only arguments
			break
		}
		ident, _ := GetParamIdent(param.X)
		return ident, param.Op
	}
	return nil, ""
}
