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
├── mode-list/           # mvd --list [SRC]
├── mode-clear/          # mvd --clear SRC
├── mode-error/          # Error handling
└── mode-dollar-expansion/ # $X env var expansion via lls config
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
| mode-add | basic-add | Add a directory to tracking |
| mode-add | add-duplicate | Adding same dir twice is idempotent |
| mode-add | add-non-existent-fails | Error when dir does not exist |
| mode-remove | basic-remove | Remove a tracked entry |
| mode-remove | remove-force | Force-remove entry with movement history |
| mode-remove | remove-no-force-with-history | Error when removing entry with history without --force |
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
