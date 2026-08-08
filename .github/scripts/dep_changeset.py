#!/usr/bin/env python3
"""Decide whether a dependency PR needs a changeset, and write one if so.

Usage: dep_changeset.py <base-ref> <pr-title> <branch-name> [force]

A dependency update only needs a release when it can reach the running image:
Go modules are compiled into the binary, and runtime JS dependencies are bundled
into the embedded UI. A bump confined to yarn.lock is build tooling, which
changes nothing a user can observe.

Prints the decision to stdout. Exits 0 whether or not a changeset was written;
the caller checks for the file.
"""

import json
import re
import subprocess
import sys
from pathlib import Path

CHANGESET_DIR = Path(".changeset")


def changed_files(base_ref):
    out = subprocess.run(
        ["git", "diff", "--name-only", f"{base_ref}...HEAD"],
        capture_output=True, text=True, check=True,
    )
    return [f for f in out.stdout.splitlines() if f]


def runtime_deps_changed(base_ref):
    """True when package.json's dependencies block differs from the base."""
    def deps_at(ref):
        r = subprocess.run(
            ["git", "show", f"{ref}:package.json"],
            capture_output=True, text=True,
        )
        if r.returncode != 0:
            return {}
        try:
            return json.loads(r.stdout).get("dependencies", {})
        except json.JSONDecodeError:
            return {}
    return deps_at(base_ref) != deps_at("HEAD")


def slug(text, limit=60):
    s = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return s[:limit] or "dependencies"


def summarise(pr_title, files):
    m = re.search(r"bump (\S+) from (\S+) to (\S+)", pr_title, re.I)
    if m:
        return f"Update {m.group(1)} from {m.group(2)} to {m.group(3)}."
    if any(f in ("go.mod", "go.sum") for f in files):
        return "Update Go dependencies."
    return "Update dependencies."


def main():
    base_ref, pr_title, branch = sys.argv[1], sys.argv[2], sys.argv[3]
    force = len(sys.argv) > 4 and sys.argv[4] == "force"

    files = changed_files(base_ref)
    go_changed = any(f in ("go.mod", "go.sum") for f in files)
    runtime_js = "package.json" in files and runtime_deps_changed(base_ref)

    if not (go_changed or runtime_js or force):
        print(f"skip: no runtime dependency change in {', '.join(files) or 'no files'}")
        return

    name = f"deps-{slug(branch)}.md"
    target = CHANGESET_DIR / name
    if target.exists():
        print(f"skip: {target} already exists")
        return

    reason = "forced by label" if force and not (go_changed or runtime_js) else (
        "go module" if go_changed else "runtime js dependency")
    CHANGESET_DIR.mkdir(exist_ok=True)
    target.write_text(
        '---\n"xbvr": patch\n---\n\n' + summarise(pr_title, files) + "\n",
        encoding="utf-8",
    )
    print(f"wrote {target} ({reason})")


if __name__ == "__main__":
    main()
