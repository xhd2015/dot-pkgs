# Scenario

**Feature**: cold walk log (P3) and incremental consume + adaptive budget (P4)

```
# cold (P3)
caller roots + CacheRoot + NoCache=false
  -> Scan cold full walk
  -> append visit events to <CacheRoot>/home/walk.jsonl
  -> on successful cold complete: append {"op":"gen_end","gen":1}
  -> write <CacheRoot>/home/walk.cursor.json with byte offset

# consume (P4)
caller after cold seal + optional FS mutations
  -> Scan warm (Refresh=false) with WarmRefreshBudget=-1
  -> Options.LastScanEnd + Options.Now select sync re-list budget
  -> consume walk.jsonl from cursor: re-list visit paths
  -> append visit|gone|… after gen_end 1; on processing gen_end G append gen_end G+1
  -> advance walk.cursor.json to new EOF

# budget tiers (P4)
delta = Now − LastScanEnd
  -> <10s: 0 sync (side/best-effort)
  -> [10s,60s): 500ms
  -> >=60s: 1s
  -> WalkConsumeSyncBudget(delta) returns the duration

# no-cache
caller roots + CacheRoot + NoCache=true
  -> Scan without walk log I/O
  -> home/walk.jsonl and home/walk.cursor.json absent
```

## Preconditions

- Nested root: own helpers; does not inherit parent tree `Setup`.
- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()` (never
  `$HOME/.cache/git-repo-scan`).
- Universe path segment is always `"home"` for these single-root library
  fixtures (aligned with P2 index serve).
- Fake `.git` fixtures only; `ListRemotes` / `ListWorktrees` off.
- Walk log schema for Assert (minimum):
  - visit: `{"op":"visit","path":"<abs-dir>"}` (extra fields allowed)
  - gone: `{"op":"gone","path":"<abs-dir>"}` (extra fields allowed)
  - seal: `{"op":"gen_end","gen":N}` (extra fields allowed)
  - cursor: `{"offset":<int>}` sealed log byte length after seal
- **P4 product Options (required):**
  - `LastScanEnd time.Time` — inject last scan end for budget; zero may read
    `home/meta.json` (`last_scan_end`) when implementer wires meta.
  - `Now func() time.Time` — already present; tests set fixed instants.
  - Export `WalkConsumeSyncBudget(sinceLast time.Duration) time.Duration`
    implementing the three tiers above.
- Consume leaves set `WarmRefreshBudget=-1` in Run so unit warm-refresh cannot
  mask or substitute walk-log consume.

## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Default `NoCache=false`, `Refresh=false`, `Debug=false`, `Consume=false`.
3. Provide `fakeGitRepo`, `absPath`, `mkdirAll` for descendants.

```go
import (
	"os"
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

// fakeGitRepo plants a minimal main-repo .git directory (objects only).
func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// homeWalkLogPath is <cacheRoot>/home/walk.jsonl
func homeWalkLogPath(cacheRoot string) string {
	return filepath.Join(cacheRoot, walkUniverseHome, "walk.jsonl")
}

// homeWalkCursorPath is <cacheRoot>/home/walk.cursor.json
func homeWalkCursorPath(cacheRoot string) string {
	return filepath.Join(cacheRoot, walkUniverseHome, "walk.cursor.json")
}

// countGenEnd returns how many gen_end events have the given gen.
func countGenEnd(events []WalkEvent, gen int) int {
	n := 0
	for _, ev := range events {
		if ev.Op == "gen_end" && ev.Gen == gen {
			n++
		}
	}
	return n
}

// lastGenEnd returns the last gen_end event and true if found.
func lastGenEnd(events []WalkEvent) (WalkEvent, bool) {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Op == "gen_end" {
			return events[i], true
		}
	}
	return WalkEvent{}, false
}

func Setup(t *testing.T, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.NoCache = false
	req.Refresh = false
	req.Debug = false
	req.ExpectWalkLog = true
	req.Consume = false
	req.BudgetOnly = false
	return nil
}
```
