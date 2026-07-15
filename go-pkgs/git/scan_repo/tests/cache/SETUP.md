# Scenario

**Feature**: per-directory mirror cache — pure store (P1), cold write (P2), warm serve (P3), budgeted refresh (P4), force refresh (P5), orphan GC (P7)

```
# pure cache store — CacheOp set, no Scan walk
caller CacheRoot + realPath + CacheOp
  -> MirrorEntryPath | SaveCacheEntry | LoadCacheEntry
  -> <CacheRoot>/mirror/<abs-without-leading-slash>/entry.json

# cold Scan write — CacheOp empty, Roots set, CacheRoot temp
caller roots + CacheRoot + NoCache -> Scan
  -> Result.Repos + optional mirror entry.json side effects

# warm Scan — CacheOp empty; Setup cold-seeds then mutates FS
caller roots + CacheRoot + NoCache=false (complete root cache)
  -> Scan serves cached live repos + liveness; NoCache bypasses warm

# warm budgeted refresh (P4) — warm-refresh/; stamp refreshed_at + YoungAge/Budget
caller warm path + WarmRefreshBudget + YoungAge (+ optional Now)
  -> rewalk oldest eligible direct-child units until budget exhausted

# force refresh (P5) — nested force-refresh/ DOCTEST (own Run with Options.Refresh)
cold seed + plant brand-new + Refresh=true
  -> cold full walk finds brand-new despite warm-eligible cache

# orphan mirror GC (P7) — orphan-gc/; parent rewalk prunes dead children
cold seed + delete real child + cold Refresh or unit rewalk
  -> mirror entry for orphan path removed (Load ok=false)
```

## Preconditions

- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()`.
- P1 leaves set `CacheOp` (mirror-path / load / save-load / overwrite).
- P2 cold-scan leaves leave `CacheOp` empty and set `Roots` under `cold-scan/`.
- P3 warm leaves leave `CacheOp` empty under `warm/`; cold seed in Setup shares
  the same `CacheRoot` as the Scan under test.
- P4 warm-refresh leaves leave `CacheOp` empty under `warm-refresh/`; stamp
  unit `refreshed_at` and set YoungAge / WarmRefreshBudget (no real 1s sleeps).
- P5 nested `force-refresh/` keeps its own DOCTEST.md; parent `Request.Refresh`
  is also used by P7 `orphan-gc/cold-rescan`.
- P7 orphan-gc leaves leave `CacheOp` empty under `orphan-gc/`; cold seed then
  delete a real child, then rewalk parent (Refresh or unit budget).
- Parse / Find paths are not exercised (`ParseURL` empty).


## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Clear enrichment / parse / find fields; leave `CacheOp` and `Roots` for descendants.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
