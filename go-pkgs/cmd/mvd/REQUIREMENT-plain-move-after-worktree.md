# Bug: Plain move after worktree resolves to worktree instead of main repo

## The Bug

When `mvd -w REPO WT` creates a worktree at WT from the main repo at REPO, the history chain becomes `[REPO, WT(worktree)]`. A subsequent plain move `mvd REPO DST` (without `-w`) should move the main repo REPO to DST. Instead, it moves the worktree WT to DST.

### Root Cause

In `resolve.go:resolveMoveSource()`, the function always returns `locations[len(locations)-1].Path` (the latest location). When the latest location happens to be a worktree (created by `mvd -w`), this returns the worktree path instead of the main repo's current location.

### Expected Behavior

For a plain move (no `-w` flag):
- If the latest location is a worktree (`Git.Type == "worktree"`), skip it and find the last non-worktree location in the chain — that is the main repo's current location.
- If the user explicitly provides the worktree path (not the root/repo name), honor that and move the worktree.
- When the main repo is moved, linked worktree `.git` files should be updated (existing behavior via `moveDir`).

## Fix Location

The fix belongs in `resolve.go`, specifically in `resolveMoveSource()`. This function is only called by `cmdMove()` (plain move, no -w). Other resolvers (`resolveBackEntry`, etc.) should NOT be affected.

### Required Logic

After resolving the latest location (`last`), check if `last.Git != nil && last.Git.Type == "worktree"`. If so, walk backwards through the locations slice to find the last non-worktree entry (the main repo's current location). Return that path as the source.

Edge cases:
- Multiple consecutive worktrees: `[repo, wt1(wt), wt2(wt)]` → should skip both to find `repo`
- Main repo was moved before worktree was created: `[repo, mid, wt(wt)]` → should find `mid`
- Explicit worktree path from user: `mvd <wt-path> dst` where wt-path is the full worktree path → should still move the worktree (current behavior is correct for this case)

The explicit worktree path case already works correctly because when the user provides a full path to the worktree and it exists:
1. `resolveBasename` returns false (file exists locally)
2. Path resolution via `resolveInputPath` finds the path
3. `findEntry` matches it in the history chain
4. `resolveMoveSource` returns it as the source

This path should be preserved.

## Test Tree

Tests are sealed (git staged) under `tests/`:

### New tests (6 RED, confirming bug):

| Test | Scenario |
|------|----------|
| `mode-move/plain-move-after-worktree` | `mvd -w repo wt` then `mvd repo dst` → should move `repo` to `dst` |
| `mode-move/plain-move-after-worktree-basename` | Same but using basename "repo" |
| `mode-move/plain-move-after-move-and-worktree` | `mvd repo mid` then `mvd -w mid wt` then `mvd repo dst` → should move `mid` to `dst` |
| `mode-move/plain-move-after-two-worktrees` | `mvd -w repo wt1` then `mvd -w repo wt2` then `mvd repo dst` → should move `repo` to `dst` |
| `mode-move/plain-move-after-multiple-moves-and-worktree` | `repo→A→B` then `-w B wt` then `mvd repo dst` → should move `B` to `dst` |
| `mode-move/plain-move-after-worktree-updates-wt-git` | After `-w repo wt` then `mvd repo dst`, wt/.git should reference dst |

### New tests (2 GREEN, sanity check):

| Test | Scenario |
|------|----------|
| `mode-move/plain-move-worktree-by-explicit-path` | `mvd <wt-path> dst` explicitly moves the worktree itself |
| `mode-back/back-remove-worktree-after-plain-move` | `--back wt` removes worktree after `repo→mid` + `-w mid wt` |

### Existing tests that must continue to pass:

All tests under `tests/mode-move/`, `tests/mode-worktree/`, `tests/mode-back/`, and all other mode directories.

## Verification

Run the new tests to confirm RED before starting:
```sh
doctest test -v ./tests/mode-move/plain-move-after-worktree
```

After implementation, verify:
```sh
doctest test -v ./tests/mode-move/plain-move-after-worktree
doctest test -v ./tests/mode-move/plain-move-after-worktree-basename
doctest test -v ./tests/mode-move/plain-move-after-move-and-worktree
doctest test -v ./tests/mode-move/plain-move-after-two-worktrees
doctest test -v ./tests/mode-move/plain-move-after-multiple-moves-and-worktree
doctest test -v ./tests/mode-move/plain-move-after-worktree-updates-wt-git
```

Also verify existing tests still pass:
```sh
doctest test -v ./tests
```

## Constraints

- Tests are sealed (git staged). Do NOT modify test files.
- If a test appears to be wrong, report back instead of modifying it.
- The fix should be minimal — only change `resolveMoveSource` in `resolve.go`.
- Do not change `resolveBackEntry`, `cmdBack`, or any other function unless necessary for the fix.
- All existing unit tests (`go test ./mvd/...`) and doctests must continue to pass.
