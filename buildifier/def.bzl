load("@bazel_skylib//lib:shell.bzl", "shell")

def _buildifier_impl(ctx):
    # That way we don't depends on defaults encoded in the binary but always
    # use defaults set on attributes of the rule
    args = [
        "-mode=%s" % ctx.attr.mode,
        "-v=%s" % str(ctx.attr.verbose).lower(),
        "-showlog=%s" % str(ctx.attr.show_log).lower(),
    ]

    if ctx.attr.lint_mode:
        args.append("-lint=%s" % ctx.attr.lint_mode)

    exclude_patterns_str = ""
    if ctx.attr.exclude_patterns:
        exclude_patterns = ["-not -path %s" % shell.quote(pattern) for pattern in ctx.attr.exclude_patterns]
        exclude_patterns_str = " ".join(exclude_patterns) + " -and"

    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    substitutions = {
        "@@ARGS@@": shell.array_literal(args),
        "@@BUILDIFIER_SHORT_PATH@@": shell.quote(ctx.executable._buildifier.short_path),
        "@@EXCLUDE_PATTERNS@@": exclude_patterns_str,
    }
    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )
    runfiles = ctx.runfiles(files = [ctx.executable._buildifier])
    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

_buildifier = rule(
    implementation = _buildifier_impl,
    attrs = {
        "verbose": attr.bool(
            doc = "Print verbose information on standard error",
        ),
        "show_log": attr.bool(
            doc = "Show log in check mode",
        ),
        "mode": attr.string(
            default = "fix",
            doc = "Formatting mode",
            values = ["check", "diff", "fix"],
        ),
        "lint_mode": attr.string(
            doc = "Linting mode",
            values = ["", "fix", "warn"],
        ),
        "exclude_patterns": attr.string_list(
            allow_empty = True,
            doc = "A list of glob patterns passed to the find command. E.g. './vendor/*' to exclude the Go vendor directory",
        ),
        "_buildifier": attr.label(
            default = "@com_github_bazelbuild_buildtools//buildifier",
            cfg = "host",
            executable = True,
        ),
        "_runner": attr.label(
            default = "@com_github_bazelbuild_buildtools//buildifier:runner.bash.template",
            allow_single_file = True,
        ),
    },
    executable = True,
)

def buildifier(**kwargs):
    tags = kwargs.get("tags", [])
    if "manual" not in tags:
        tags.append("manual")
        kwargs["tags"] = tags
    _buildifier(
        **kwargs
    )
