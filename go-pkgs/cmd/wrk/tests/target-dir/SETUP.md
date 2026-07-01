# Scenario

**Feature**: wrk <dir> <target-dir> overrides the worktree spawn location

```
# second positional <target-dir> overrides default {WRK_HOME}/worktrees spawn path
myrepo (main) -> wrk myrepo <target-dir> -> worktree at <target-dir> or <target-dir>/<name>
# <target-dir> resolved relative to shell cwd (process cwd), NOT relative to <dir>
wrk <dir> <target-dir> -> spawn path overridden; WRK_HOME ignored
# create-only: <target-dir> + --list/--done/--dep -> wrk: unexpected arguments
```

## Preconditions

- Git must be available for all leaves (every leaf spawns or rejects a worktree).
- Source repo `myrepo` lives on branch `main` with one commit; `WRK_DATE=2026-06-30`.

## Steps

- Every leaf initializes the source repo `myrepo` on `main` under `{WorkRoot}`.
- Leaves run `wrk <myrepo> <target-dir> [flags...]` from process cwd `{WorkRoot}` (so a
  relative `<target-dir>` resolves against the shell cwd, not the repo dir).
- `req.TargetDir` = `{WorkRoot}/myrepo` (absolute source repo, first positional).
- `req.SpawnDir` = the `<target-dir>` under test (absolute for most leaves; relative for
  `relative-path/`). `req.Args` carries any trailing flags (`--list`, `--dep <dep>`).
- Expected worktree paths are NOT under `{WRK_HOME}/worktrees`; assert funcs compute them
  directly with `filepath.Join(req.WorkRoot, ...)`.

## Context

- The worktree spawn location changes from `{WRK_HOME}/worktrees/...` to either
  `<target-dir>` (target missing, parent exists) or
  `<target-dir>/{basename}-{token}-{WRK_DATE}[-N]` (target exists). Branch naming is
  unchanged. `WRK_HOME` is ignored when `<target-dir>` is given.
- basename=`myrepo`, token=`main`, date=`2026-06-30` for all leaves.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, repoDir)

	req.TargetDir = repoDir
	req.RepoDir = req.WorkRoot
	return nil
}
```
