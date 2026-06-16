# mvd Test Cases

Decision tree covering all `mvd` commands and their behaviors.

## Tree Overview

```
mvd tests
├── mode-move/           # mvd SRC DST (default move)
├── mode-worktree/       # mvd -w SRC DST (git worktree)
├── mode-add/            # mvd --add DIR
├── mode-remove/         # mvd --rm DIR
├── mode-rebase/         # mvd --rebase DIR NEW-DIR
├── mode-back/           # mvd --back SRC
├── mode-list/           # mvd --list [SRC] + --picker-list (marker tests)
├── mode-clear/          # mvd --clear SRC
├── mode-error/          # Error handling
├── mode-dollar-expansion/ # $X env var expansion via lls config
├── mode-alias-storage/    # aliases stored inside history.json
├── mode-dry-run/          # --dry-run flag (skips modifications, prints intent)
├── mode-safety/           # overlapping paths between moves and worktrees
└── mode-worktree-back-enhanced/  # enhanced --back for worktrees (CASE B: ff-merge prompt, CASE C: rebase)
```

## Test Case Index

| Mode | Leaf | Description |
|------|------|-------------|
| mode-move | basic-move | Move src → dst (flat target) |
| mode-move | move-to-existing-dir | Move into an existing directory (basename join) |
| mode-move | multi-step-move | Two sequential moves forming a chain |
| mode-move | move-as-root-path | Move using the original root path as SRC |
| mode-move | move-by-basename | Move using a unique root basename |
| mode-move | move-by-alias | Move using a registered alias |
| mode-move | ambiguous-basename | Error when basename matches multiple roots |
| mode-move | plain-move-after-worktree | After `-w REPO WT`, plain move `REPO DST` moves the main repo (not WT) |
| mode-move | plain-move-after-worktree-basename | Same as above but using basename resolution |
| mode-move | plain-move-worktree-by-explicit-path | Explicit worktree path still moves the worktree itself |
| mode-move | plain-move-after-move-and-worktree | After `REPO→MID` + `-w MID WT`, plain move `REPO DST` moves MID to DST |
| mode-move | plain-move-after-two-worktrees | After two `-w` calls, plain move skips both worktrees to find main repo |
| mode-move | plain-move-after-multiple-moves-and-worktree | Deep chain: multiple moves + worktree; plain move finds main repo |
| mode-move | plain-move-after-worktree-updates-wt-git | Plain move updates worktree .git file to new main repo location |
| mode-worktree | worktree-move | Create worktree with -w flag |
| mode-worktree | worktree-non-git-src | Error when SRC is not a git repo |
| mode-worktree | worktree-back-dirty | Error when worktree has uncommitted changes |
| mode-worktree | worktree-back-unmerged | Error when worktree branch is unmerged |
| mode-worktree | worktree-back-success | Successful worktree back after merge |
| mode-worktree | worktree-branch-collision | Branch name collision generates date-suffixed name |
| mode-worktree | worktree-move-by-basename | Worktree creation using basename |
| mode-worktree | move-worktree-with-w-flag | Explicit -w flag uses git worktree add |
| mode-worktree | move-worktree-without-w-flag | Without -w, worktree is moved via os.Rename |
| mode-worktree | move-nested-worktree-without-w-flag | Nested worktree .git file is updated |
| mode-worktree | worktree-move-to-existing-dir | Worktree creation when destination is an existing directory |
| mode-worktree | worktree-by-alias | Worktree creation using a registered alias |
| mode-worktree | worktree-alias-not-found | Error when alias is not registered for -w |
| mode-add | basic-add | Add a directory to tracking |
| mode-add | add-duplicate | Adding same dir twice is idempotent |
| mode-add | add-non-existent-fails | Error when dir does not exist |
| mode-remove | basic-remove | Remove a tracked entry |
| mode-remove | remove-force | Force-remove entry with movement history |
| mode-remove | remove-no-force-with-history | Error when removing entry with history without --force |
| mode-remove | remove-by-chain-path | Remove a non-root path from a chain, preserving root |
| mode-remove | remove-worktree-entry | Remove one worktree entry from a multi-worktree chain |
| mode-rebase | basic-rebase | Rebase entry to a new root |
| mode-rebase | rebase-by-basename | Rebase using basename resolution |
| mode-back | basic-back | Move back one step |
| mode-back | back-after-plain-move-which-followed-worktree | `--back` on moved repo skips worktree entries to find correct prev location |
| mode-back | back-at-origin | Back at origin is a no-op |
| mode-back | back-by-basename | Back using unique basename |
| mode-back | back-by-alias | Back using a registered alias (bug fix) |
| mode-back | multi-step-back | Multi-step back to origin |
| mode-back | back-remove-worktree-after-plain-move | `--back WT` removes worktree after `REPO→MID` + `-w MID WT` |
| mode-list | list-all | List all tracked projects |
| mode-list | list-single | List a single project's history |
| mode-list | list-by-basename | List by basename |
| mode-list | list-picker-root-plus-worktree | Picker dump shows root + 1 worktree (2 entries) |
| mode-list | list-picker-two-worktrees | Picker dump shows root + 2 worktrees (3 entries) |
| mode-list | list-picker-plain-move | Picker dump for plain move shows only latest (1 entry) |
| mode-list | list-picker-after-back | Picker dump after --back shows only root (1 entry) |
| mode-list | list-picker-alias-with-worktree | Alias annotation on root entry, not worktree |
| mode-list | marker-worktree-basic | Root `(main)` + worktree `(worktree)` markers, alive, no alias |
| mode-list | marker-worktree-two | Root `(main)` + 2 worktrees `(worktree)` markers, alive |
| mode-list | marker-external-main | External main path shown `(external main)` — bug fix for root→WT→plain→WT |
| mode-list | marker-external-main-is-latest | External main that is also latest — not duplicated |
| mode-list | marker-alias-with-main | Combined marker `(main, aliases: ...)` when root has alias + is main |
| mode-list | marker-alias-no-worktree | `(aliases: ...)` on plain entry without worktree |
| mode-list | marker-dead-worktree | `(dead worktree)` for dead worktree path |
| mode-list | marker-dead-main | `(dead main)` for dead root that is also main |
| mode-list | marker-dead-external-main | `(dead external main)` for dead external main path |
| mode-list | marker-dead-main-with-alias | Combined `(dead main, aliases: ...)` for dead root with alias |
| mode-list | marker-dead-plain | `(dead)` for dead plain entry (no worktree) |
| mode-list | marker-no-marker | No marker for plain alive entry without worktree or alias |
| mode-clear | basic-clear | Clear history for a project |
| mode-clear | clear-by-basename | Clear by basename |
| mode-error | non-existent-src | Error when SRC does not exist |
| mode-error | move-non-existent-basename | Error when basename matches nothing |
| mode-dollar-expansion | add-with-dollar | --add with $X/myproject |
| mode-dollar-expansion | back-with-dollar | --back with $X/myproject |
| mode-dollar-expansion | clear-with-dollar | --clear with $X/myproject |
| mode-dollar-expansion | list-with-dollar | --list with $X/myproject |
| mode-dollar-expansion | move-default-with-dollar | Move with $X/myproject |
| mode-dollar-expansion | rebase-with-dollar | --rebase with $X/myproject |
| mode-dollar-expansion | which-with-dollar | --which with $X/myproject |
| mode-dollar-expansion | worktree-move-with-dollar | -w with $X/myrepo |
| mode-alias-storage | add-alias-not-creates-aliases-file | --add-alias does not create aliases.json; alias stored in history.json |
| mode-alias-storage | add-alias-survives-history-save-load | Alias survives history save/load cycle after another move |
| mode-alias-storage | multiple-aliases-per-project | Multiple aliases for same project stored in history.json |
| mode-dry-run | dry-run-move | --dry-run with plain move: prints "would move", skips os.Rename + history write |
| mode-dry-run | dry-run-move-to-dir | --dry-run move into existing dir: basename join, no actual move |
| mode-dry-run | dry-run-worktree | --dry-run with -w: prints "would create worktree", skips git worktree add |
| mode-dry-run | dry-run-add | --dry-run with --add: prints "would add", skips history write |
| mode-dry-run | dry-run-add-alias | --dry-run with --add-alias: prints "would add alias", alias not persisted |
| mode-dry-run | dry-run-rm | --dry-run with --rm: prints "would remove", history entry retained |
| mode-dry-run | dry-run-rm-force | --dry-run with --rm -f: force path exercised, history entry retained |
| mode-dry-run | dry-run-rebase | --dry-run with --rebase: prints "would rebase", history unchanged |
| mode-dry-run | dry-run-back | --dry-run with --back (plain): prints "would move back", no os.Rename |
| mode-dry-run | dry-run-back-worktree | --dry-run with --back (worktree): prints "would remove worktree", no git worktree remove |
| mode-dry-run | dry-run-back-at-origin | --dry-run --back at origin: "nothing to move back", no dry-run message |
| mode-dry-run | dry-run-clear | --dry-run with --clear: prints "would clear", history intact |
| mode-dry-run | dry-run-cd | --dry-run with --cd: prints "would cd", no shell launched |
| mode-dry-run | dry-run-vscode | --dry-run with --vscode: prints "would open VSCode", no code launched |
| mode-dry-run | dry-run-error-nosrc | --dry-run with non-existent SRC: validation error still fires, no dry-run message |
| mode-dry-run | dry-run-error-non-git | --dry-run -w with non-git SRC: validation error still fires |
| mode-dry-run | dry-run-list | --dry-run with --list: read-only command runs normally, no dry-run output |
| mode-dry-run | dry-run-which | --dry-run with --which: read-only command runs normally, no dry-run output |
| mode-dry-run | dry-run-picker-list | --dry-run with --picker-list: read-only command runs normally, no dry-run output |
| mode-safety | move-parent-with-worktree | Move parent dir containing tracked repo + WT; WT .git stays stale (BUG) |
| mode-safety | back-after-parent-move | --back parent after Scenario A; sub-project history is dead (BUG) |
| mode-safety | back-worktree-stale-mainrepo | --back WT after main repo moved; position mismatch (BUG: stale MainRepo) |
| mode-safety | move-into-worktree-dir | Plain move of main repo into its own worktree directory |
| mode-safety | back-from-nested-in-worktree | --back restores from nested position; WT .git updated correctly |
| mode-safety | move-to-existing-worktree-path | Plain move targeting path that IS an existing worktree (joins basename) |
| mode-safety | back-long-chain-worktree-middle | Long chain; --back skips WT entry for prev |
| mode-worktree-back-enhanced | dirty-worktree | Dirty worktree → error (existing behavior unchanged) |
| mode-worktree-back-enhanced | branch-merged | Branch already merged → success (existing behavior unchanged) |
| mode-worktree-back-enhanced | branch-ahead/confirm-default | HEAD ancestor of branch, user presses Enter → ff merge + remove |
| mode-worktree-back-enhanced | branch-ahead/decline | HEAD ancestor of branch, user types 'n' → abort, no changes |
| mode-worktree-back-enhanced | branch-ahead/non-tty | HEAD ancestor of branch, stdin not a TTY → error |
| mode-worktree-back-enhanced | branches-diverged/rebase-success | Neither ancestor, confirm (Enter) → rebase+ff merge+remove |
| mode-worktree-back-enhanced | branches-diverged/rebase-conflict | Neither ancestor, confirm (Enter) → rebase conflicts → abort rebase, error |
| mode-worktree-back-enhanced | branches-diverged/decline | Neither ancestor, decline ('n') → abort, no changes |
| mode-worktree-back-enhanced | branches-diverged/non-tty | Neither ancestor, no TTY → error |
| mode-worktree-back-enhanced | back-at/diverged-rebase-splice | cmdWorktreeBackAt: diverged, confirm (Enter) → rebase succeeds → splice chain |
| mode-worktree-back-enhanced | back-at/ahead-confirm-splice | cmdWorktreeBackAt: branch ahead, confirm → splice chain |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Run all tests
doctest test ./tests

# Run a specific mode
doctest test ./tests/mode-dry-run
```
