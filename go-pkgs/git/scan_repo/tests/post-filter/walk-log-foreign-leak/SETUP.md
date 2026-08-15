# Scenario

**Bug**: warm walk-log consume re-lists a foreign parent visit and leaks
`agent-pro` into `Result.Repos` even though scan root is only the consumer.

```
# layout under temp base
worktrees/consumer/                 # scan root — main checkout (fake .git)
worktrees/other/external/agent-pro/ # foreign main — NOT under consumer

# seed
cold Scan(Roots=[consumer]) → home/repos.json warm-eligible for consumer
inject visit abs(worktrees/other/external) before gen_end in walk.jsonl
ancient last_scan_end so WalkConsumeSyncBudget → 1s

# under test
Scan(Roots=[consumer], CacheRoot, WarmRefreshBudget=-1, ListWorktrees=false)
  -> consume re-lists foreign parent → would liveRepoAt(agent-pro)
  -> Result.Repos must contain only paths under consumer
  -> no foreign agent-pro path
```

## Preconditions

- Isolated `CacheRoot` (parent Setup).
- `ListWorktrees` false (progress + top-level filter only; no expand).
- `WarmRefreshBudget=-1` so discoveries cannot come from unit rewalk.
- Ancient `LastScanEnd` + fixed `Now` → delta ≥ 60s → full consume budget.
- Classic TDD: RED while consume merges foreign checkouts without base filter.

## Steps

1. Create `worktrees/consumer` and `worktrees/other/external/agent-pro` with fake `.git`.
2. Cold-seed Scan against consumer only.
3. Inject walk visit of foreign parent `…/other/external` before last `gen_end`.
4. Stamp ancient last_scan_end clocks; disable unit warm-refresh.
5. Stash `ConsumerPath` / `ForeignPath` for Assert.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	base := t.TempDir()
	consumer := filepath.Join(base, "worktrees", "consumer")
	foreignParent := filepath.Join(base, "worktrees", "other", "external")
	foreign := filepath.Join(foreignParent, "agent-pro")
	mkdirAll(t, consumer)
	mkdirAll(t, foreign)
	fakeGitRepo(t, consumer)
	fakeGitRepo(t, foreign)

	consumerAbs := absPath(t, consumer)
	foreignAbs := absPath(t, foreign)
	foreignParentAbs := absPath(t, foreignParent)

	req.Roots = []string{consumerAbs}
	req.NoCache = false
	req.Refresh = false
	req.ListWorktrees = false
	req.ListRemotes = false
	// Isolate walk-log consume from budgeted unit rewalk.
	req.WarmRefreshBudget = -1

	coldSeedScan(t, req.Roots, req.CacheRoot)

	// Foreign parent visit: re-list discovers agent-pro as a child checkout.
	injectVisitBeforeLastGenEnd(t, req.CacheRoot, foreignParentAbs)

	// Ancient last_scan_end → full 1s walk-consume budget (delta ≥ 60s).
	t0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	req.LastScanEnd = t0
	req.SetLastScanEnd = true
	req.NowAt = now
	req.SetNow = true
	// Also stamp meta so product paths that ignore Options still age correctly.
	if err := scan_repo.SaveLastScanEnd(req.CacheRoot, t0); err != nil {
		t.Fatalf("SaveLastScanEnd: %v", err)
	}

	req.ConsumerPath = consumerAbs
	req.ForeignPath = foreignAbs
	return nil
}
```
