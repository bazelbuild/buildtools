"""Provides go_yacc and genfile_check_test

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
"""

_GO_YACC_TOOL = "@org_golang_x_tools//cmd/goyacc"

def go_yacc(src, out, visibility=None):
  """Runs go tool yacc -o $out $src."""
  native.genrule(
      name = src + ".go_yacc",
      srcs = [src],
      outs = [out],
      tools = [_GO_YACC_TOOL],
      cmd = ("export GOROOT=$$(dirname $(location " + _GO_YACC_TOOL + "))/..;" +
             " $(location " + _GO_YACC_TOOL + ") " +
             " -o $(location " + out + ") $(SRCS)"),
      visibility = visibility,
      local = 1,
  )
      
