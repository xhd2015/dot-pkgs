# Scenario

**Feature**: durable repo index + walk JSONL + budgets (dense mirror retired)

```
# cold Scan write — CacheOp empty, Roots set, CacheRoot temp
caller roots + CacheRoot + NoCache -> Scan
  -> Result.Repos + home/repos.json + walk.jsonl/gen_end
  -> no <CacheRoot>/mirror

# warm Scan — CacheOp empty; Setup cold-seeds then mutates FS
caller roots + CacheRoot + NoCache=false (usable index)
  -> Scan serves index + live repos + liveness + optional sibling probe
  -> NoCache bypasses warm; walk-log consume under adaptive budget

# warm budgeted unit refresh — warm-refresh/; unit dir ModTime + YoungAge/Budget
caller warm path + WarmRefreshBudget + YoungAge (+ optional Now)
  -> rewalk oldest eligible direct-child units until budget exhausted

# force refresh — nested force-refresh/ DOCTEST (own Run with Options.Refresh)
cold seed + plant brand-new + Refresh=true
  -> cold full walk finds brand-new despite warm-eligible index

# no-mirror/ — RED until product stops writing mirror/
cold/warm Scan with CacheRoot -> mirror path absent

# debug timing/mode (nested debug/ DOCTEST — Options.Debug + stderr capture)
Debug=true cold/warm → scan: mode= + serve timing; Debug=false → zero scan:

# nested trees (own DOCTEST.md each)
repo-index/   → Load/SaveRepoIndex + ApplyLiveness (home|root)
index-serve/  → Scan seeds/serves home/repos.json + sibling + liveness
walk-log/     → walk.jsonl cold seal gen_end + consume + adaptive budget
```

## Preconditions

- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()`.
- Cold-scan / warm / warm-refresh / no-mirror leave `CacheOp` empty and set `Roots`.
- Warm leaves leave `CacheOp` empty under `warm/`; cold seed in Setup shares
  the same `CacheRoot` as the Scan under test.
- Warm-refresh leaves leave `CacheOp` empty under `warm-refresh/`; stamp
  unit ModTime and set YoungAge / WarmRefreshBudget (no real 1s sleeps).
- Nested `force-refresh/` keeps its own DOCTEST.md.
- Nested `debug/`, `repo-index/`, `index-serve/`, `walk-log/` each keep their
  own DOCTEST.md (inheritance firewall).
- Parse / Find paths are not exercised (`ParseURL` empty).
- Pure mirror store (`MirrorEntryPath` / `SaveCacheEntry` / `LoadCacheEntry`) is retired.

## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Clear enrichment / parse / find fields; leave `CacheOp` and `Roots` for descendants.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.ListRemotes = false
	req.ListWorktrees = false
	req.ParseURL = ""
	req.FindGitHubOwner = ""
	req.FindGitHubRepo = ""
	req.Refresh = false
	return nil
}
```
