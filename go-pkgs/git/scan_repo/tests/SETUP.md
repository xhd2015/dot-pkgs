# Scenario

**Feature**: `scan_repo.Scan` discovers git repos; `ParseRemoteOwnerRepo` parses remote URLs; durable index + walk log (dense mirror retired)

```
# discovery pipeline
caller roots + options -> Scan -> Walk -> repo rows (sorted by Path)

# enrichment (optional)
Scan -> git subprocesses -> Remotes[] / Worktrees[] on main rows

# URL parser (standalone)
ParseRemoteOwnerRepo(url) -> owner, repo, ok

# durable cache (v2 — no dense mirror/)
caller roots + CacheRoot + NoCache=false -> Scan
  -> Result.Repos + home/repos.json + home/walk.jsonl
  -> <CacheRoot>/mirror MUST NOT be written

# warm Scan serve + liveness
cold seed (home/repos.json) -> optional FS mutate
  -> Scan(NoCache=false, CacheRoot) serves from index + .git liveness
  -> NoCache=true always full live walk

# warm budgeted refresh (P4)
warm path + WarmRefreshBudget + YoungAge + unit dir ModTime
  -> rewalk oldest eligible direct-child units; merge new repos into Result

# force refresh (P5) — nested tests/cache/force-refresh/
cold seed -> plant brand-new -> Scan(Refresh=true)
  -> full cold walk finds brand-new (warm would omit)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo` is importable.
- `Scan` uses filesystem walk only when enrichment flags are false.
- Enrichment branches require `git` on PATH.
- Cache branches always use an explicit temp `CacheRoot` (never `$HOME/.cache`).
- Cold-scan leaves pass `CacheRoot` into `Options` and assert index via `LoadRepoIndex`.
- Warm leaves cold-seed in Setup then assert the second Scan via `Run`.
- Warm-refresh leaves stamp unit ModTimes and set budget options (no real 1s sleeps).
- P5 nested `cache/force-refresh/` keeps its own DOCTEST.md.
- Dense `mirror/` / `MirrorEntryPath` / `SaveCacheEntry` / `LoadCacheEntry` are **retired**.

## Context

- Paths in assertions use `filepath.Abs` and `filepath.Clean` for portability.
- `fakeGitRepo` / `fakeGitWorktree` avoid git for pure discovery tests.
- Real git helpers skip the test when `git` is unavailable.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}
```
