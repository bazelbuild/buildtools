"""
This module is the public interface to buildifier rules
"""

load(
    ":buildifier.bzl",
    _buildifier = "buildifier",
    _buildifier_test = "buildifier_test",
)

buildifier = _buildifier
buildifier_test = _buildifier_test
