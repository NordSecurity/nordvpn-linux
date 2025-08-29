#!/usr/bin/env python3
"""
Detect whether changes are GUI-only or include core (non-GUI) files.

Outputs (to stdout and $GITHUB_OUTPUT):
  gui_changed=true|false
  core_changed=true|false

Usage:
  python detect_changes.py --gui-dir gui
"""

from __future__ import annotations

import argparse
import contextlib
import json
import os
import subprocess
import sys
from collections.abc import Iterable


def sh(*cmd: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, text=True, capture_output=True, check=check)


def git_has_commit(sha: str) -> bool:
    if not sha:
        return False
    try:
        sh("git", "rev-parse", "--verify", f"{sha}^{{commit}}")
        return True
    except subprocess.CalledProcessError:
        return False


def git_fetch_commit(sha: str) -> None:
    if not sha:
        return
    # Fetch that single object if it’s not present (handles shallow/rewrites)
    if not git_has_commit(sha):
        with contextlib.suppress(subprocess.CalledProcessError):
            sh("git", "fetch", "--no-tags", "--depth=1", "origin", sha, check=True)


def determine_base_head(
    event_name: str,
    event_path: str,
    head_sha: str,
) -> tuple[str | None, str]:
    """Return (BASE, HEAD). BASE may be None for new branches / first commit."""
    with open(event_path, encoding="utf-8") as f:  # UP015: omit default "r"
        payload = json.load(f)

    if event_name == "pull_request":
        base = payload.get("pull_request", {}).get("base", {}).get("sha")
        head = payload.get("pull_request", {}).get("head", {}).get("sha") or head_sha
        return base, head

    # push / tag / others
    base = payload.get("before") or ""
    head = head_sha

    # First push on a new branch can have empty/zero "before"
    if not base or base == "0000000000000000000000000000000000000000":
        # Try parent of HEAD; may not exist on initial commit
        try:
            base = sh("git", "rev-list", "-n", "1", f"{head}^").stdout.strip()
        except subprocess.CalledProcessError:
            base = ""
    return (base or None), head


def changed_files(base: str | None, head: str) -> Iterable[str]:
    """Yield paths changed between BASE..HEAD (or just HEAD if BASE is None)."""
    if base:
        # Ensure both commits exist locally
        git_fetch_commit(base)
        git_fetch_commit(head)
        try:
            out = sh("git", "diff", "--name-only", f"{base}..{head}").stdout
        except subprocess.CalledProcessError:
            out = ""
    else:
        # Single-commit range (HEAD^!); if no parent, fall back to show-files
        git_fetch_commit(head)
        try:
            out = sh("git", "diff", "--name-only", f"{head}^!").stdout
        except subprocess.CalledProcessError:
            try:
                out = sh("git", "show", "--name-only", "--pretty=", head).stdout
            except subprocess.CalledProcessError:
                out = ""

    for line in out.splitlines():
        p = line.strip()
        if p:
            # Normalize leading "./"
            if p.startswith("./"):
                p = p[2:]
            yield p


def classify_changes(paths: Iterable[str], gui_dir: str) -> tuple[bool, bool]:
    """Return (gui_changed, core_changed)."""
    gui_prefix = f"{gui_dir.rstrip('/')}/"
    gui_changed = False
    core_changed = False

    for p in paths:
        if p.startswith(gui_prefix):
            gui_changed = True
        else:
            core_changed = True

        # Short-circuit if we already know both changed
        if gui_changed and core_changed:
            break

    return gui_changed, core_changed


def write_output(gui_changed: bool, core_changed: bool) -> None:
    line_gui = f"gui_changed={'true' if gui_changed else 'false'}"
    line_core = f"core_changed={'true' if core_changed else 'false'}"
    print(line_gui)
    print(line_core)

    out_path = os.environ.get("GITHUB_OUTPUT")
    if out_path:
        with open(out_path, "a", encoding="utf-8") as fh:
            fh.write(f"{line_gui}\n{line_core}\n")


def main() -> int:
    parser = argparse.ArgumentParser(description="Detect GUI-only vs core changes")
    parser.add_argument("--gui-dir", default="gui", help="Root directory of GUI code (default: gui)")
    args = parser.parse_args()

    # Tag pushes: treat as “everything changed”
    ref = os.environ.get("GITHUB_REF", "")
    if ref.startswith("refs/tags/"):
        write_output(gui_changed=True, core_changed=True)
        return 0

    event_name = os.environ.get("GITHUB_EVENT_NAME", "")
    event_path = os.environ.get("GITHUB_EVENT_PATH", "")
    head_sha = os.environ.get("GITHUB_SHA", "")

    if not event_name or not event_path or not head_sha:
        print("Required GitHub env is missing; assuming both changed.", file=sys.stderr)
        write_output(gui_changed=True, core_changed=True)
        return 0

    base, head = determine_base_head(event_name, event_path, head_sha)
    paths = list(changed_files(base, head))
    gui_changed, core_changed = classify_changes(paths, args.gui_dir)
    write_output(gui_changed, core_changed)
    return 0


if __name__ == "__main__":
    sys.exit(main())
