#!/usr/bin/env python3
"""Migrate ~/.mvd/history.json from v2.0 (prev/current/type) to v3.0 (from/to/from_type/to_type).

Usage:
  python3 migrate_history_v3.py [path/to/history.json]

Backs up the original file as history.json.bak before writing.
"""

from __future__ import annotations

import json
import shutil
import sys
from pathlib import Path


def location_type(path: str, worktree_paths: set[str]) -> str:
    return "worktree" if path in worktree_paths else "main"


def migrate_move_v2(old: dict, chain_types: dict[str, str]) -> dict:
    from_path = old["prev"]
    to_path = old["current"]
    op_type = old.get("type", "plain")

    if op_type == "worktree":
        to_type = "worktree"
    else:
        to_type = "main"

    from_type = chain_types.get(from_path, "main")

    move = {
        "from": from_path,
        "from_type": from_type,
        "to": to_path,
        "to_type": to_type,
    }
    if old.get("branch"):
        move["branch"] = old["branch"]

    chain_types[to_path] = to_type
    return move


def migrate_move_v3(move: dict) -> dict:
    required = ("from", "from_type", "to", "to_type")
    if all(k in move for k in required):
        return move
    raise ValueError(f"unrecognized move entry: {move!r}")


def migrate_project(proj: dict) -> dict:
    root = proj.get("root")
    if not root:
        return proj

    moves = proj.get("moves") or []
    if not moves:
        return proj

    # Already v3.0
    if "from" in moves[0] and "to" in moves[0]:
        return proj

    chain_types: dict[str, str] = {root: "main"}
    new_moves = []
    for old in moves:
        if "prev" in old and "current" in old:
            new_moves.append(migrate_move_v2(old, chain_types))
        else:
            new_moves.append(migrate_move_v3(old))

    return {**proj, "moves": new_moves}


def migrate(data: dict) -> dict:
    version = data.get("version", "")
    projects = data.get("projects")
    if not projects:
        return data

    if version == "3.0":
        print("already v3.0, nothing to do")
        return data

    migrated = {key: migrate_project(proj) for key, proj in projects.items()}
    return {"version": "3.0", "projects": migrated}


def main() -> int:
    default = Path.home() / ".mvd" / "history.json"
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else default

    if not path.exists():
        print(f"error: {path} does not exist", file=sys.stderr)
        return 1

    with path.open() as f:
        data = json.load(f)

    backup = path.with_suffix(path.suffix + ".bak")
    shutil.copy2(path, backup)
    print(f"backup: {backup}")

    out = migrate(data)
    with path.open("w") as f:
        json.dump(out, f, indent=2)
        f.write("\n")

    print(f"migrated: {path} -> version 3.0")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())