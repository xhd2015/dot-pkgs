# Bug: --back fails after plain move that followed worktree creation

## The Bug

Given this sequence:
1. `mvd -w base target` — creates a worktree at `target` from `base`
2. `mvd base another` — moves `base` to `another`, updates `target`'s `.git` to point to `another`
3. `mvd --back another` — **ERROR**: `rename .../another .../target: file exists`

The `--back` command tries to move `another` back to `target` (the worktree directory which still exists) instead of moving it back to `base`.

## Root Cause

In `mv.go:cmdBack()` at line 82:
```go
prev := locations[len(locations)-2]
```

This picks the positional previous entry in the `locations` array. When the chain is `[base, target(wt), another]`, `prev` resolves to `target` (the worktree) instead of `base`. The code should skip worktree entries when finding the previous non-worktree location to move back to.

## Data Model Migration (Already Done)

The history file format has been migrated from v1.1 to v2.0. The new format records explicit move pairs:

**v2.0 format** (`main.go` already updated):
```json
{
  "version": "2.0",
  "projects": {
    "/path/to/repo": {
      "root": "/path/to/repo",
      "moves": [
        {"prev": "/path/to/repo", "current": "/path/to/wt", "type": "worktree", "branch": "wt"},
        {"prev": "/path/to/repo", "current": "/path/to/dst", "type": "plain"}
      ]
    }
  }
}
```

- `MoveEntry` type added to `main.go` (Prev, Current, Type, Branch)
- `deriveMoves(locs)` converts locations → moves (for saving)
- `locationsFromMoves(root, moves)` converts moves → locations (for loading)
- `saveHistory` writes v2.0 format with `root` + `moves`
- `loadHistory` reads v2.0 format and converts back to `History` (map[string][]LocationEntry) for internal backward compat

The `agent-pro` entry in `~/.mvd/history.json` has been migrated correctly.

## Required Fix

The `--back` operation must use explicit move records instead of positional inference. There are two levels of fix possible:

### Minimal Fix: Fix `cmdBack` to skip worktree entries in `prev`

In `mv.go:cmdBack()`, when `last` is NOT a worktree but `prev` (locations[len(locations)-2]) IS a worktree, walk backwards through locations to find the last non-worktree entry as the true `prev` destination. This mirrors the existing `findLastNonWorktreePath` helper in `resolve.go`.

### Robust Fix: Use moves model for --back

Refactor `cmdBack` to work with the moves array:
- Find the last move for the project
- If type == "worktree": call `cmdWorktreeBack`
- If type == "plain": `moveDir(move.Current, move.Prev)`, then remove the move from history
- Update `cmdMove` to append moves correctly (already handled by `deriveMoves` on save)

Either approach must produce the correct behavior for the test case.

## Test Tree

**New test leaf** (sealed, RED):

| Test | Path | Description |
|------|------|-------------|
| `back-after-plain-move-which-followed-worktree` | `tests/mode-back/back-after-plain-move-which-followed-worktree/` | After `-w` then plain move, `--back` on the moved dir returns it to original location |

**Setup**: Create git repo → `mvd -w base target` → `mvd base another` → `mvd --back another`
**Assert**: Exit 0, "moved back" in output, `another` gone, `base` exists with README.md, `target` still exists (worktree intact), history chain = `[base, target(wt)]`.

## Verification

Run the new test to confirm it passes (currently RED):
```sh
doctest test -v ./mvd/tests/mode-back/back-after-plain-move-which-followed-worktree
```

Run all tests to confirm no regressions (only one pre-existing failure: worktree-move-by-basename):
```sh
doctest test -v ./mvd/tests
```

Also run unit tests:
```sh
go test ./mvd/...
```

## Constraints

- Tests are sealed (git staged). Do NOT modify test files under `tests/`.
- The test root `SETUP.md` has been staged with updated v2.0-aware helpers.
- `main.go` has unstaged changes (v2.0 format types, saveHistory, loadHistory). These are the data model migration — you may modify main.go further if needed but the v2.0 format must be preserved.
- If a test assertion appears wrong, report back instead of modifying it.
- All existing doctests (except pre-existing `worktree-move-by-basename`) must continue to pass.
- Keep changes minimal and focused on `mv.go:cmdBack()` and related functions.
