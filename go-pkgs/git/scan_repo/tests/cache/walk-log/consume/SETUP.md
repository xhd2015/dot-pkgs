# Scenario

**Feature**: second Scan after cold seal consumes walk.jsonl under adaptive budget

```
# two-phase protocol
cold Scan -> gen_end 1 + cursor at EOF
  -> optional: delete visit path | plant new repo
  -> second Scan(Refresh=false, WarmRefreshBudget=-1,
                 LastScanEnd, Now)
  -> re-list from cursor within WalkConsumeSyncBudget(delta)
  -> append events after gen_end 1; seal gen_end 2 when gen_end 1 consumed
  -> cursor advances to new EOF
```

## Preconditions

- Grouping for all P4 consume / budget leaves.
- `req.Consume=true` unless a child sets `BudgetOnly=true`.
- Default clock pair for full-budget consume: LastScanEnd = T0, Now = T0+2m
  (delta ≥ 60s → 1s sync budget) unless a budget leaf overrides.
- Warm unit refresh is disabled in Run (`WarmRefreshBudget=-1`).

## Steps

1. Set `Consume=true`, `NoCache=false`, `ExpectWalkLog=true`.
2. Default `SetLastScanEnd` + `SetNow` with delta ≥ 60s for non-budget leaves.
3. Leaves plant fixtures and may override clock / mutations / BudgetOnly.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoCache = false
	req.Refresh = false
	req.ExpectWalkLog = true
	req.Consume = true
	req.BudgetOnly = false

	// Default: ample delta so consume has 1s sync budget (product tiers).
	t0 := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	req.LastScanEnd = t0
	req.SetLastScanEnd = true
	req.NowAt = t0.Add(2 * time.Minute) // delta = 120s >= 60s → 1s
	req.SetNow = true
	req.DeltaAge = 2 * time.Minute
	return nil
}
```
