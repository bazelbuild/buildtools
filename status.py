#!/usr/bin/env python

from __future__ import print_function
from subprocess import Popen, PIPE
from sys import exit


def run(*cmd):
    process = Popen(cmd, stdout=PIPE)
    output, err = process.communicate()
    if process.wait():
        exit(1)
    return output.strip()


def main():
    tags = run("git", "describe", "--tags")
    print("STABLE_buildVersion {}".format(tags.split("-")[0]))

    revision = run("git", "rev-parse", "HEAD")
    print("STABLE_buildScmRevision {}".format(revision))


if __name__ == "__main__":
    main()
