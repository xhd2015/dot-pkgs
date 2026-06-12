# Feature: Merge aliases into history.json

## Summary

Move alias storage from standalone `~/.mvd/aliases.json` into `~/.mvd/history.json`. Aliases become per-project data inside `history.json`.

## What changes

### 1. `history/types.go` — Add `Aliases` field to `ProjectEntry`

```go
type ProjectEntry struct {
    Root      string          `json:"root,omitempty"`
    Locations []LocationEntry `json:"locations,omitempty"`
    Moves     []MoveEntry     `json:"moves,omitempty"`
    Aliases   []string        `json:"aliases,omitempty"`
}
```

### 2. `history/io.go` — Update Save/Load

- **`Save(path string, hist History, aliases map[string]string) error`**: Accept aliases map, write per-project aliases arrays in history.json. For each project in hist, look up all aliases pointing to that root and populate the Aliases field.
- **`Load(path string) (History, map[string]string, error)`**: Extract aliases from project entries and return them alongside History. Reconstruct `map[string]string` by mapping each alias name to its project root.

### 3. `main.go` — Integrate alias storage with history

- Remove `aliasesPath()` function entirely
- Remove `loadAliases()` — aliases now come from `loadHistory()` (which calls `history.Load()` returning both history and aliases)
- Remove `saveAliases()` — aliases now saved via `saveHistory()` (which calls `history.Save()` with both history and aliases)
- Update `cmdAddAlias()`: add alias to the aliases map, then save via saveHistory
- Update all callers of `loadAliases()` to use the new combined load function. These are in: `main.go` (multiple functions), `cd.go`, `open.go`, `mv.go`, `picker.go`, `resolve.go`

### 4. One-time migration script

Write a Python script at `scripts/migrate_aliases.py` that:
- Reads `~/.mvd/aliases.json` 
- Reads `~/.mvd/history.json`
- For each alias in aliases.json, finds the corresponding project entry in history.json and adds the alias name to its `aliases` array
- If the project doesn't exist in history, creates a minimal project entry with just the alias
- Saves the updated history.json
- Renames `~/.mvd/aliases.json` to `~/.mvd/aliases.json.bak`

### 5. `scripts/migrate_history.py` (existing)

Update if needed to handle the new aliases field.

## Test Tree (sealed — DO NOT MODIFY)

All new tests under `tests/mode-alias-storage/`:

```
mode-alias-storage/
├── SETUP.md
├── add-alias-not-creates-aliases-file/
│   ├── SETUP.md    — adds project, adds alias via --add-alias
│   └── ASSERT.md   — alias in history.json; aliases.json does NOT exist
├── add-alias-survives-history-save-load/
│   ├── SETUP.md    — adds 2 projects, alias for 1st, moves 2nd (triggers save/load)
│   └── ASSERT.md   — alias "mp" still in history.json; aliases.json does NOT exist
└── multiple-aliases-per-project/
    ├── SETUP.md    — adds project, adds two aliases ("mp", "myproj-alias")
    └── ASSERT.md   — both aliases in history.json; both shown in --list; aliases.json does NOT exist
```

### What each leaf asserts:
1. `add-alias-not-creates-aliases-file`: `aliases.json` does NOT exist, alias "mp" found in history.json project entry's `aliases` array
2. `add-alias-survives-history-save-load`: After another project is moved (triggering history save/load), alias "mp" is still in history.json, aliases.json does NOT exist
3. `multiple-aliases-per-project`: Both "mp" and "myproj-alias" in history.json and `--list` output, aliases.json does NOT exist

### Existing tests that may need attention:
- `mode-list/list-picker-alias-with-worktree/SETUP.md` — uses `writeAliasesFile()` to write aliases.json. This test must be updated since mvd no longer reads aliases.json. The test setup should instead write aliases into history.json's project entry. This is a **necessary test update**, not an arbitrary test modification. The assertion in ASSERT.md should remain unchanged (it checks picker-dump output contains alias annotation).

## Verification command

```sh
doctest test -v ./tests
```

All tests must pass. The 3 new tests must pass. All existing tests must continue to pass. The only allowed test modification is updating `list-picker-alias-with-worktree/SETUP.md` to write aliases into history.json instead of aliases.json.

## Design Notes

- No backward compatibility with standalone aliases.json needed after migration
- The one-time migration is a Python script, not built into mvd
- Aliases are stored per-project: each ProjectEntry has `"aliases": ["name1", "name2"]` 
- The aliases map (alias name → root path) is reconstructed from per-project aliases on load
- When saving, for each project root, collect all aliases pointing to that root and write them
