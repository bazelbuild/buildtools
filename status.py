#!/usr/bin/env python

from __future__ import print_function, unicode_literals
from subprocess import Popen, PIPE
from sys import exit


def run(*cmd):
    process = Popen(cmd, stdout=PIPE)
    output, err = process.communicate()
    if process.wait():
        exit(1)
    return output.strip().decode()


def main():
    tags = run("git", "describe", "--tags")
    print("STABLE_buildVersion", tags.split("-")[0])

    # rules_nodejs expects to read from volatile-status.txt
    print("BUILD_SCM_VERSION", tags.split("-")[0])
    
    revision = run("git", "rev-parse", "HEAD")
    print("STABLE_buildScmRevision", revision)


if __name__ == "__main__":
    main()
