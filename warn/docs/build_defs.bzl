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
