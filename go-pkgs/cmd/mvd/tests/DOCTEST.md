# mvd Test Cases

## Version
0.0.2

Decision tree covering all `mvd` commands and their behaviors.

# DSN (Domain Specific Notion)

- **mvd CLI** — move/rename tool for projects, git worktrees, and history tracking.
- **Config home** — isolated `MVD_DEBUG_CONFIG_HOME` per test for `history.json`.
- **Work root** — temp directory holding repos and move targets.

## Tree Overview

```
mvd tests
├── mode-move/           # mvd SRC DST (default move)
├── mode-worktree/       # mvd -w SRC DST (git worktree)
├── mode-add/            # mvd --add DIR
├── mode-remove/         # mvd --rm DIR
├── mode-history-v3/     # history.json v3.0 moves schema
├── mode-rebase/         # mvd --rebase DIR NEW-DIR
├── mode-back/           # mvd --back SRC
├── mode-list/           # mvd --list [SRC] + --picker-list (marker tests)
├── mode-clear/          # mvd --clear SRC
├── mode-error/          # Error handling
├── mode-dollar-expansion/ # $X env var expansion via lls config
├── mode-alias-storage/    # aliases stored inside history.json
├── mode-dry-run/          # --dry-run flag (skips modifications, prints intent)
├── mode-safety/           # overlapping paths between moves and worktrees
└── mode-worktree-back-enhanced/  # enhanced --back for worktrees (CASE B/C; default auto-yes; --confirm restores prompts)
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
| mode-worktree | worktree-spawn-from-worktree | Spawn worktree with -w when SRC is a linked worktree |
| mode-worktree | worktree-spawn-from-worktree-dry-run | --dry-run -w with linked worktree SRC prints intent, skips creation |
| mode-worktree | worktree-non-git-src | Error when SRC is not a git repo |
| mode-worktree | worktree-back-dirty | Error when worktree has uncommitted changes |
| mode-worktree | worktree-back-unmerged | Ahead (unmerged) non-TTY `--back` auto-yes succeeds (ff-merge + remove) |
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
| mode-remove | remove-dead-external-main | `--rm` on dead external main chain path (findEntry, not hist[absDir]) |
| mode-remove | remove-dead-worktree-chain-path | `--rm` on dead worktree in long chain; root and siblings preserved |
| mode-history-v3 | save-format-after-move | After add + worktree + plain move, history.json is v3.0 moves format |
| mode-history-v3 | external-main-worktree-move | v3.0 moves round-trip + picker `(external main)` for agent-pro topology |
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
| mode-list | marker-moved-worktree | BUG: plain-move of worktree loses worktree marker (shows external main) |
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
| mode-error | grep-empty | `--grep ''` → non-zero; requires non-empty pattern |
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
| mode-dry-run | dry-run-back-worktree | --dry-run with --back (worktree): lists planned git -C commands + "would remove worktree", no mutations |
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
| mode-worktree-back-enhanced | branch-ahead/pipe-without-confirm-flag | Piped stdin without confirm flags → auto-yes ff-merge + remove |
| mode-worktree-back-enhanced | branch-ahead/confirm-default | Bare `--back` auto-yes: HEAD ancestor → ff merge + remove (no Proceed?) |
| mode-worktree-back-enhanced | branch-ahead/decline | `--confirm` + `--confirm-from-stdin` + `n` → abort, no changes |
| mode-worktree-back-enhanced | branch-ahead/prompt-shows-commands | `--confirm` + decline: FormatPlanPrompt lists git -C merge/remove/delete commands |
| mode-worktree-back-enhanced | branch-ahead/non-tty | HEAD ancestor, non-TTY bare `--back` → auto-yes success (no Proceed?) |
| mode-worktree-back-enhanced | branches-diverged/rebase-success | Bare `--back` auto-yes: neither ancestor → rebase+ff merge+remove |
| mode-worktree-back-enhanced | branches-diverged/rebase-conflict | Bare `--back` auto-yes → rebase conflicts → abort rebase, error |
| mode-worktree-back-enhanced | branches-diverged/decline | `--confirm` + `--confirm-from-stdin` + `n` → abort, no changes |
| mode-worktree-back-enhanced | branches-diverged/prompt-shows-commands | `--confirm` + decline: FormatPlanPrompt lists git -C rebase/merge/remove/delete commands |
| mode-worktree-back-enhanced | branches-diverged/non-tty | Neither ancestor, non-TTY bare `--back` → auto-yes rebase+merge success |
| mode-worktree-back-enhanced | back-at/diverged-rebase-splice | cmdWorktreeBackAt: diverged bare auto-yes → rebase succeeds → splice chain |
| mode-worktree-back-enhanced | back-at/ahead-confirm-splice | cmdWorktreeBackAt: branch ahead bare auto-yes → splice chain |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Run all tests
doctest test ./tests

# Run a specific mode
doctest test ./tests/mode-dry-run
```

```go
import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	ConfigHome string
	WorkRoot   string
	Args       []string
	StdinInput string
	UseScript  bool
}

type Response struct {
	Output   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	bin := getMvdBin(t, d)

	var cmdArgs []string
	var cmdName string
	if req.UseScript && req.StdinInput == "" {
		cmdName = "script"
		cmdArgs = append([]string{"-q", "/dev/null", bin}, req.Args...)
	} else {
		cmdName = bin
		cmdArgs = req.Args
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = append(os.Environ(), "MVD_DEBUG_CONFIG_HOME="+req.ConfigHome)

	if req.StdinInput != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if req.UseScript {
			time.Sleep(200 * time.Millisecond)
		}
		io.WriteString(stdin, req.StdinInput)
		stdin.Close()
		err = cmd.Wait()
		exitCode := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			} else {
				return nil, err
			}
		}
		return &Response{Output: outBuf.String(), ExitCode: exitCode}, nil
	}

	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Output: string(out), ExitCode: exitCode}, nil
}
```
