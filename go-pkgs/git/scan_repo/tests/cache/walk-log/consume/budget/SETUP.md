# Scenario

**Feature**: adaptive walk-consume sync budget from delta(Now, LastScanEnd)

```
# tiers (product WalkConsumeSyncBudget)
delta < 10s        -> 0          # side/best-effort; tests assert 0 sync discover
10s <= delta < 60s -> 500ms
delta >= 60s       -> 1s

# Options
LastScanEnd time.Time  # inject or read meta
Now func() time.Time   # inject fixed clock
```

## Preconditions

- Grouping for budget-tier leaves only.
- Pure selection leaves use `BudgetOnly=true` (no Scan) where that isolates
  the tier table; `delta-lt-10s` also runs a behavioral Consume path to prove
  zero sync discover.
- Documented thresholds are exclusive of the next band at the 10s and 60s
  boundaries: `[0,10) → 0`, `[10s,60s) → 500ms`, `[60s,∞) → 1s`.

## Steps

1. Mark this as the budget split group (no fixture plant here).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Budget group: keep cache on; children set BudgetOnly / DeltaAge / clocks.
	req.NoCache = false
	req.Refresh = false
	req.Debug = false
	return nil
}
```
