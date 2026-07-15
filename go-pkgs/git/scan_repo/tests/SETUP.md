# Scenario

**Feature**: `scan_repo.Scan` discovers git repos; `ParseRemoteOwnerRepo` parses remote URLs; cache store maps and persists per-dir entries

```
# discovery pipeline
caller roots + options -> Scan -> Walk -> repo rows (sorted by Path)

# enrichment (optional)
Scan -> git subprocesses -> Remotes[] / Worktrees[] on main rows

# URL parser (standalone)
ParseRemoteOwnerRepo(url) -> owner, repo, ok

# mirror cache store (P1 pure I/O)
caller CacheRoot + realPath -> MirrorEntryPath / SaveCacheEntry / LoadCacheEntry
  -> <CacheRoot>/mirror/<abs-without-leading-slash>/entry.json

# cold Scan cache write (P2 side effect)
caller roots + CacheRoot + NoCache=false -> Scan (full walk)
  -> Result.Repos (discovery unchanged) + mirror entry.json for visited dirs

# warm Scan serve + liveness (P3)
cold seed (complete root cache) -> optional FS mutate
  -> Scan(NoCache=false, CacheRoot) serves from mirror is_repo + .git liveness
  -> NoCache=true always full live walk

# warm budgeted refresh (P4)
warm path + WarmRefreshBudget + YoungAge + stamped refreshed_at
  -> rewalk oldest eligible direct-child units; merge new repos into Result

# force refresh (P5) — nested tests/cache/force-refresh/
cold seed -> plant brand-new -> Scan(Refresh=true)
  -> full cold walk finds brand-new (warm would omit)

# orphan mirror GC (P7) — tests/cache/orphan-gc/
cold seed -> delete real child -> parent rewalk (Refresh or unit refresh)
  -> dead child mirror entry removed (Load ok=false)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo` is importable.
- `Scan` uses filesystem walk only when enrichment flags are false.
- Enrichment branches require `git` on PATH.
- Cache branches always use an explicit temp `CacheRoot` (never `$HOME/.cache`).
- Cold-scan leaves pass `CacheRoot` into `Options` and assert mirror via `LoadCacheEntry`.
- Warm leaves cold-seed in Setup then assert the second Scan via `Run`.
- Warm-refresh leaves stamp unit ages and set budget options (no real 1s sleeps).
- P5 nested `cache/force-refresh/` keeps its own DOCTEST.md; parent also passes
  `Refresh` for P7 orphan-gc cold-rescan.
- P7 orphan-gc leaves assert stronger prune than P3 liveness (`entry` gone).



## Context

- Paths in assertions use `filepath.Abs` and `filepath.Clean` for portability.
- `fakeGitRepo` / `fakeGitWorktree` avoid git for pure discovery tests.
- Real git helpers skip the test when `git` is unavailable.
- Cache helpers (`expectedMirrorEntryPath`) encode the P1 path-mapping spec for assertions and seed fixtures.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// expectedMirrorEntryPath implements the P1 mapping spec used by cache leaves:
// Abs+Clean realPath, strip leading separator, join cacheRoot/mirror/.../entry.json.
func expectedMirrorEntryPath(t *testing.T, cacheRoot, realPath string) string {
	t.Helper()
	abs := absPath(t, realPath)
	rel := strings.TrimPrefix(abs, string(filepath.Separator))
	if rel == "" || rel == abs {
		// still under root separator only, or non-Unix volume path: keep cleaned abs
		// without a leading separator when present
		rel = strings.TrimPrefix(abs, string(filepath.Separator))
	}
	return filepath.Join(cacheRoot, "mirror", rel, "entry.json")
}
```
