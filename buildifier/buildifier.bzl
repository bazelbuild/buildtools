"""
The module defines buildifier as a Bazel rule.
"""

load(
    "//buildifier/internal:factory.bzl",
    "buildifier_attr_factory",
    "buildifier_impl_factory",
)

def _buildifier_impl(ctx):
    return [buildifier_impl_factory(ctx)]

_buildifier = rule(
    implementation = _buildifier_impl,
    attrs = buildifier_attr_factory(),
    executable = True,
)

def buildifier(**kwargs):
    """
    Wrapper for the _buildifier rule. Adds 'manual' to the tags.

    Args:
      **kwargs: all parameters for _buildifier
    """

    tags = kwargs.get("tags", [])
    if "manual" not in tags:
        tags.append("manual")
        kwargs["tags"] = tags
    _buildifier(**kwargs)

def _buildifier_test_impl(ctx):
    return [buildifier_impl_factory(ctx, test_rule = True)]

buildifier_test = rule(
    implementation = _buildifier_test_impl,
    attrs = buildifier_attr_factory(True),
    test = True,
)
