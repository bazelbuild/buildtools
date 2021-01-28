"""Provides documentation rule

Copyright 2021 Google Inc. All Rights Reserved.

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

def _documentation_impl(ctx):
    ctx.actions.run(
        executable = ctx.executable.bin,
        inputs = [ctx.files.textproto[0]],
        outputs = [ctx.outputs.markdown],
        arguments = [ctx.files.textproto[0].path, ctx.outputs.markdown.path],
    )
    return DefaultInfo(
        files = depset([ctx.outputs.markdown]),
    )

documentation = rule(
    implementation = _documentation_impl,
    attrs = {
        "textproto": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "markdown": attr.output(),
        "bin": attr.label(
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
    },
)
