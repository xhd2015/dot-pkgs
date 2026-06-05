#!/usr/bin/env python3
"""
Migrate ~/.mvd/history.json from the legacy flat format to the v1.1 format.

Legacy format (map[string][]string):
    {"/orig/path": ["/orig/path", "/moved/path"]}

New format (v1.1):
    {
      "version": "1.1",
      "projects": {
        "/orig/path": {
          "locations": [
            {"path": "/orig/path"},
            {"path": "/moved/path"}
          ]
        }
      }
    }

This script backs up the original file to .bak (skip if already exists),
then writes the converted file.
"""

import json
import os
import sys

HISTORY_PATH = os.path.expanduser("~/.mvd/history.json")


def legacy_to_v11(data: dict) -> dict:
    projects = {}
    for key, locs in data.items():
        if not isinstance(locs, list):
            print(f"  warning: skipping invalid entry for key {key!r}", file=sys.stderr)
            continue
        projects[key] = {
            "locations": [{"path": loc} for loc in locs]
        }
    return {"version": "1.1", "projects": projects}


def is_v11(data: dict) -> bool:
    return "version" in data and "projects" in data


def main():
    if not os.path.isfile(HISTORY_PATH):
        print(f"{HISTORY_PATH} does not exist, nothing to migrate")
        return

    with open(HISTORY_PATH, "r") as f:
        data = json.load(f)

    if is_v11(data):
        print(f"{HISTORY_PATH} is already in v1.1 format, nothing to do")
        return

    backup_path = HISTORY_PATH + ".bak"
    if os.path.isfile(backup_path):
        print(f"backup already exists: {backup_path}, skipping backup")
    else:
        with open(backup_path, "w") as f:
            json.dump(data, f, indent=2)
        print(f"backed up to {backup_path}")

    converted = legacy_to_v11(data)
    os.makedirs(os.path.dirname(HISTORY_PATH), exist_ok=True)
    with open(HISTORY_PATH, "w") as f:
        json.dump(converted, f, indent=2)
    print(f"migrated {HISTORY_PATH} to v1.1 ({len(converted['projects'])} projects)")


if __name__ == "__main__":
    main()
