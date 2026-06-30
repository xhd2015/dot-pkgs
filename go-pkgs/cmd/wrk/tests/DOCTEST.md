# wrk Test Cases

## Version
0.0.2

Decision tree covering the `wrk` CLI: no-args worktree creation, optional `wrk <dir>`
first positional, `wrk --dep` external dependency worktrees, `wrk --done` merge-back
(including external cascade), and `wrk --list`.

# DSN (Domain Specific Notion)

- **wrk CLI** — standalone binary; invocation form `wrk [dir] [flags...]`; first non-flag positional argument is optional `<dir>` — when present, effective cwd is the resolved absolute path of `<dir>` (process cwd unchanged); when absent, effective cwd is process cwd; no-args invocation creates a git worktree from the effective checkout and prints the target path on stdout; `wrk --done` merges the linked worktree branch back into main and removes the worktree + branch (`worktree.MergeBack` with `Remove: true`); `wrk --list` runs `git -C <effective-cwd> worktree list` and prints stdout unchanged; missing `<dir>` → `wrk: <path> does not exist`.
- **WRK_HOME** — storage root env var (default `~/.wrk`); tests isolate per run at `{WorkRoot}/.wrk`.
- **WRK_DATE** — optional env var (`YYYY-MM-DD`) overriding the run date for deterministic naming; all tests set `WRK_DATE=2026-06-30`.
- **Work root** — temp directory holding source repos and move targets.
- **Naming** — worktree path `{WRK_HOME}/worktrees/{basename}-{token}-{YYYY-MM-DD}[-N]`; branch `{base-branch}-{YYYY-MM-DD}[-N]`; `N` starts at 1 on collision (path exists or branch ref exists). No unsuffixed names without date.
- **token** — `sanitize(base-branch)` for normal branches (`/` → `-` in path only); for detached HEAD, 7-char short commit hash from `git rev-parse --short=7 HEAD` (not literal `HEAD`).
- **Git source** — cwd must be inside a git checkout (main repo, linked worktree, or nested subdirectory); basename resolves from the checkout root when cwd is a linked worktree or nested subpath.
- **wrk --dep** — spawns a dependency worktree under `{consumerTop}/external/` via `git worktree add`, appends `/external` to `.gitignore` when missing, runs `gotool.Replace` + `gotool.Tidy`, prints external worktree abs path on stdout. Naming: `{basename}-{token}-{WRK_DATE}[-N]` (same rules as create; basename from dep main repo).
- **wrk --done** — resolves checkout root via `ShowToplevel(cwd)`; requires a linked worktree (not main repo); clean worktree; implicit `--rm`. **Cascade**: merge-back each linked worktree under `{toplevel}/external/*` first. **Guard**: error if consumer `go.mod` has any filesystem/local `replace` (`./`, `../`, or absolute path without version). Branch relation to main: already-included → remove only; ahead/diverged → prompt then merge/rebase.
- **wrk --list** — runs `git -C <cwd> worktree list`; prints stdout unchanged; cwd must be inside a git work tree (main repo, linked worktree, or nested subpath). Mutually exclusive with no-args create and `--done`.
- **--confirm-from-stdin** — when set with piped `StdinInput`, reads Y/n from stdin for merge-back confirmation (required for non-TTY ahead/diverged cases).
- **Request.Args** — CLI arguments passed to `wrk` after optional `<dir>` (empty → no-args create; `["--dep", depPath]` for dep tests; `["--done"]` or `["--done", "--confirm-from-stdin"]` for done tests; `["--list"]` for list tests).
- **Request.TargetDir** — when set, prepended as the first positional argument to `wrk` (`wrk <dir> ...`); used by `dir-arg/` tests to run from `WorkRoot` while targeting a repo elsewhere.
- **external/** — dependency worktrees live at `{consumerTop}/external/{basename}-{token}-{WRK_DATE}[-N]`; not under `WRK_HOME`.
- **gitWorktreeList** — helper capturing raw `git worktree list` stdout from a directory for list-test comparison.
- **Request.StdinInput** — when non-empty, piped to wrk stdin before wait (mvd merge-back pattern).

## Tree Overview

```
wrk tests
├── create-worktree/              # cwd is a git checkout (success path)
│   ├── main-checkout/            # cwd is the main repo checkout
│   │   ├── basic-create/         # first wrk from main
│   │   ├── sequence-increment/   # second wrk increments -N suffix
│   │   ├── branch-collision/     # branch ref blocks date-suffixed name
│   │   └── slash-branch/         # branch with / sanitized in path token
│   ├── from-linked-worktree/     # cwd is linked worktree; basename from main repo
│   ├── from-git-subpath/         # cwd is nested subdir inside checkout; basename from repo root
│   └── detached-head/            # cwd on detached HEAD → 7-char hash token
├── dep/                          # wrk --dep external dependency worktree
│   ├── basic/                    # require + --dep → external wt, replace, tidy, gitignore
│   ├── gitignore-already/        # /external already in .gitignore → no duplicate
│   ├── not-a-dependency/         # dep not in go.mod → error
│   ├── not-git-repo/             # dep path not git → error
│   └── not-go-module/            # dep git without go.mod → error
├── done/                         # wrk --done merge-back --rm from linked worktree
│   ├── already-included/         # branch already merged into main; remove only
│   ├── ahead-confirm/            # ahead + --confirm-from-stdin + Enter
│   ├── ahead-decline/            # ahead + --confirm-from-stdin + decline
│   ├── ahead-non-tty/            # ahead without confirm flag (non-interactive)
│   ├── dirty/                    # uncommitted changes in worktree
│   ├── not-linked/               # cwd is main repo (not linked worktree)
│   ├── from-subpath/             # cwd nested inside linked wt; uses checkout root
│   ├── external-cascade/         # cascade removes external/* wt; guard blocks parent
│   └── local-replace-blocks/     # filesystem replace in go.mod → error at guard
├── list/                         # wrk --list (git worktree list wrapper)
│   ├── main-only/                # single main checkout, no linked worktrees
│   ├── with-linked/              # main + one linked worktree
│   ├── from-subpath/             # cwd nested inside main repo
│   └── non-git/                  # cwd is not a git repo (error)
├── non-git-cwd/                  # cwd is not a git repo (error, no-args create)
└── dir-arg/                      # wrk <dir> optional first positional
    ├── create/
    │   └── basic/                # wrk <repoDir> from WorkRoot creates worktree
    ├── list/
    │   └── from-dir/             # wrk <repoDir> --list from WorkRoot
    └── missing-dir/              # wrk <nonexistent> → does not exist
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | create-worktree/main-checkout/basic-create | Fresh git repo on `main`, first `wrk` |
| 2 | create-worktree/main-checkout/sequence-increment | Run `wrk` twice from same repo/branch |
| 3 | create-worktree/main-checkout/branch-collision | Pre-create branch `main-{date}`, no worktree at date path |
| 4 | non-git-cwd | cwd is not a git repository |
| 5 | create-worktree/from-linked-worktree | cwd is existing linked worktree |
| 6 | create-worktree/detached-head | cwd on detached HEAD uses 7-char hash token |
| 7 | create-worktree/main-checkout/slash-branch | Branch `feature/foo` |
| 8 | create-worktree/from-git-subpath | cwd is nested subdir inside checkout; basename from repo root |
| 9 | done/already-included | wt branch merged into main; `--done` removes wt + branch |
| 10 | done/ahead-confirm | wt ahead; `--done --confirm-from-stdin` + `\n` → ff-merge + remove |
| 11 | done/ahead-decline | wt ahead; `--confirm-from-stdin` + `n\n` → aborted, wt remains |
| 12 | done/ahead-non-tty | wt ahead; no confirm flag → non-zero (cannot prompt) |
| 13 | done/dirty | uncommitted file in wt → non-zero |
| 14 | done/not-linked | cwd is main repo → `not a linked worktree` |
| 15 | done/from-subpath | cwd is subdir inside linked wt; `--done` uses checkout root |
| 16 | list/main-only | single main repo; stdout matches `git worktree list` |
| 17 | list/with-linked | main + linked worktree; stdout lists both paths |
| 18 | list/from-subpath | cwd nested in main repo; stdout same as from repo root |
| 19 | list/non-git | cwd is not a git repository |
| 20 | dep/basic | Consumer requires dep; `--dep` creates external wt, replace, tidy, gitignore |
| 21 | dep/gitignore-already | `/external` already present → no duplicate line |
| 22 | dep/not-a-dependency | Dep not in go.mod → error |
| 23 | dep/not-git-repo | Path not git → error |
| 24 | dep/not-go-module | Git dir without go.mod → error |
| 25 | done/external-cascade | `--done` cascades to `external/*` dep wt first; parent errors on local replace |
| 26 | done/local-replace-blocks | `replace => ./external/foo` blocks `--done` at guard |
| 27 | dir-arg/create/basic | `wrk <repoDir>` from WorkRoot creates worktree |
| 28 | dir-arg/list/from-dir | `wrk <repoDir> --list` matches `git worktree list` |
| 29 | dir-arg/missing-dir | `wrk <nonexistent>` → does not exist |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Run all tests (expect RED until wrk is implemented)
doctest test ./tests

# Run a specific leaf
doctest test ./tests/create-worktree/main-checkout/basic-create

# Run a done leaf
doctest test ./tests/done/ahead-confirm

# Run a list leaf
doctest test ./tests/list/main-only

# Run a dep leaf
doctest test ./tests/dep/basic

# Run a done cascade leaf
doctest test ./tests/done/external-cascade

# Run a dir-arg leaf
doctest test ./tests/dir-arg/create/basic
doctest test ./tests/dir-arg/list/from-dir
doctest test ./tests/dir-arg/missing-dir
```

```go
import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot   string
	WrkHome    string
	RepoDir    string   // process cwd when running wrk
	TargetDir  string   // optional first positional <dir>; prepended to Args when set
	HashToken  string   // detached-head: 7-char short commit hash
	Args       []string // CLI args after <dir>; empty → no-args create
	StdinInput string   // piped to stdin when set
	MainRepo      string // done tests: main checkout path
	WtDir         string // done/dep tests: linked worktree path
	WtBranch      string // done tests: worktree branch name
	DepPath       string // dep tests: path-to-repo argument
	ConsumerTop   string // dep tests: consumer git toplevel
	ExternalWtDir string // dep/done tests: external worktree path
	DepModulePath string // dep tests: module path from dep go.mod
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)

	var args []string
	if req.TargetDir != "" {
		args = append(args, req.TargetDir)
	}
	args = append(args, req.Args...)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)

	if req.StdinInput != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(stdin, req.StdinInput); err != nil {
			return nil, err
		}
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
		return &Response{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
		}, nil
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```