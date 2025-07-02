#!/usr/bin/env python3

from sh import git, go
import sys
import os
import tempfile
import argparse
import json
import random
import string

patchset_go = """
package main

import (
	"fmt"
	"encoding/json"
	"github.com/openrelayxyz/xplugeth"
)

func main() {
	x, err := json.Marshal(xplugeth.GetPatchsets())
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(x))
}"""


def getPatches(cmd):
    orig = os.getcwd()
    os.chdir(cmd)
    try:
        with open("get_patchset.go", "w") as fd:
            fd.write(patchset_go)
        x = json.loads(go.run("-tags=patchset xplugeth", "get_patchset.go", "xplugeth_imports.go"))
        os.remove("get_patchset.go")
        return x
    finally:
        os.chdir(orig)

def apply_patches(patches, tags):
    for patchset in patches:
        apply_patchset(patchset, tags)

def apply_patchset(patchset, tags):
    start_ref = git("rev-parse", "HEAD").strip()
    for patch in patchset:
        try:
            apply_patch(patch, tags)
        except Exception as e:
            git.reset(start_ref, "--hard")
            print(e)
            continue
        else:
            break
    else:
        raise Exception("No successful pathces applied for patchset")

def apply_patch(patch, tags):
    remote_name = "".join(random.choice(string.ascii_lowercase) for _ in range(6))
    git.remote("add", remote_name, sshRemoteTransformer(patch["remote"]))
    print(git.fetch(remote_name))
    print(git("cherry-pick", patch["ref"]))
    for test in patch["tests"]:
        for testName in test["test"]:
            t = ','.join(tags.split())
            print(test["package"], f"-tags={t}", "-run", testName)
            print(go.test(test["package"], "-run", testName))

def parse_source(remote):
    if remote.lower().strip("/") == 'https://github.com/ethereum/go-ethereum':
        return 'foundation'
    elif remote.lower().strip("/") == 'https://github.com/maticnetwork/bor':
        return 'bor'
    elif remote.lower().strip("/") == 'https://github.com/etclabscore/core-geth':
        return 'etc'
    else:
        return remote.split("/")[-1]


def push_to_archive(archive, remote, tag, xplugeth_tag, xplugeth_branch):
    try:
        git.remote.add("archive", archive)
    except Exception as e:
        print(f"encountered an exception adding archive remote: {e}")

    source = parse_source(remote)
    branch = xplugeth_tag + "-" + xplugeth_branch + "-" + source + "-" + tag

    git.checkout('-b', branch)
    git.push(archive, f'HEAD:{branch}')


def main(remote, tag, plugins, cmd, artifacts_directory, workdir, replacements, archive, build_tags):
    if archive:
        xp_branch = git("rev-parse", "--abbrev-ref", "HEAD").strip()
        xp_tag = git("describe", "--tags", "--abbrev=0").strip()


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
        git.reset("HEAD", "--hard")
        git.clean("-fdx")
        git.checkout(tag)
        with open("go.mod", "a") as fd:
            for package, local in replacements:
                fd.write(f"\n replace {package} => {local}")
        with open(os.path.join(cmd, "xplugeth_imports.go"), "w") as fd:
            fd.write("//go:build xplugeth\npackage main\nimport (\n")
            for plugin in plugins:
                print(go.get(plugin))
                fd.write('\t_ "%s"\n' % (plugin.split("@")[0]))
            fd.write(")")
        git.add("go.mod")
        git.add("go.sum")
        git.add(os.path.join(cmd, "xplugeth_imports.go"))
        git.config("user.name", "xplugeth-build")
        git.config("user.email", "build@plugeth.org")
        git.commit("-m", "xplugeth-build: add plugin imports")

        apply_patches(getPatches(cmd), build_tags)

        t = ','.join('xplugeth'.split() + build_tags.split())
        print(go.build(f"-tags={t}", "-o", os.path.join(artifacts_directory, os.path.split(cmd)[-1]), cmd))
    finally:
        if archive:
            try:
                push_to_archive(archive, remote, tag, xp_tag, xp_branch)
            except Exception as e:
                print(f"error pushing to archive remote: {e}")
        os.chdir(orig)


def sshRemoteTransformer(repo):
    return "git@" + ":".join(repo.split("/", maxsplit=1))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
                    prog='xplugeth',
                    description='Build extended Geth binaries')
    parser.add_argument('-s', '--source-remote', default="https://github.com/ethereum/go-ethereum") 
    parser.add_argument('-t', '--source-tag', default="v1.15.3")
    parser.add_argument('-p', '--plugin', action="append", default=[])
    parser.add_argument('-r', '--replace', action="append", default=[])
    parser.add_argument('-c', '--cmd', default="./cmd/geth")
    parser.add_argument('-w', '--workdir', default=None)
    parser.add_argument('-a', '--artifacts-directory', default="/tmp/output/")
    parser.add_argument('-v', '--archive', nargs="?", const="git@github.com:openrelayxyz/xplugeth-archive.git", default=None)
    parser.add_argument('-b', '--build-tags', default=None)
    # note: the archive url needs to be ssh to preserve users git credentials

    args = parser.parse_args()

    for item in args.replace:
        if '=' not in item:
            print(f"arguments to replace must contain an '='")
            sys.exit()
    replacements = [replace.split("=") for replace in args.replace]

    if args.workdir:
        main(args.source_remote, args.source_tag, args.plugin or ["github.com/openrelayxyz/xplugeth/build"], args.cmd, args.artifacts_directory, args.workdir, replacements, args.archive, args.build_tags)
    else:
        with tempfile.TemporaryDirectory() as workdir:
            main(args.source_remote, args.source_tag, args.plugin or ["github.com/openrelayxyz/xplugeth/build"], args.cmd, args.artifacts_directory, workdir, replacements, args.archive, args.build_tags)
                    