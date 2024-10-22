#!/usr/bin/env python3

from sh import git, go
import sys
import os
import tempfile
import argparse

def main(remote, tag, plugins, cmd, artifacts_directory, workdir):
    if not os.path.exists(os.path.join(workdir, ".git")):
        git.clone(remote, workdir)
    else:
        try:
            git.remote.add("project", remote)
        except Exception:
            pass
        git.fetch("project")
    orig = os.getcwd()
    try:
        os.chdir(workdir)
        git.checkout(tag)
        with open(os.path.join(cmd, "xplugeth_imports.go"), "w") as fd:
            fd.write("package main\nimport (\n")
            for plugin in plugins:
                go.get(plugin)
                fd.write('\t_ "%s"\n' % (plugin.split("@")[0]))
            fd.write(")")

        print(go.build("-o", os.path.join(artifacts_directory, os.path.split(cmd)[-1]), cmd))
    finally:
        os.chdir(orig)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
                    prog='xplugeth',
                    description='Build extended Geth binaries')
    parser.add_argument('-s', '--source-remote', default="https://github.com/ethereum/go-ethereum") 
    parser.add_argument('-t', '--source-tag', default="v1.14.7")
    parser.add_argument('-p', '--plugin', action="append", default=[])
    parser.add_argument('-c', '--cmd', default="./cmd/geth")
    parser.add_argument('-w', '--workdir', default=None)
    parser.add_argument('-a', '--artifacts-directory', default="/tmp/output/")

    args = parser.parse_args()
    if args.workdir:
        main(args.source_remote, args.source_tag, args.plugin, args.cmd, args.artifacts_directory, args.workdir)
    else:
        with tempfile.TemporaryDirectory() as workdir:
            main(args.source_remote, args.source_tag, args.plugin, args.cmd, args.artifacts_directory, workdir)
            