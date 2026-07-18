# Scenario

**Feature**: delta &lt; 10s selects 0 sync budget — no walk-consume discover of new repo

```
# fresh re-scan soon after last_scan_end
cold: workspace with projects/alpha only; gen_end 1
  -> plant projects/beta after cold
  -> second Scan with LastScanEnd=T0, Now=T0+5s (delta=5s < 10s)
  -> WarmRefreshBudget=-1 (no unit refresh)
  -> WalkConsumeSyncBudget(5s) == 0
  -> Result must NOT include beta (0 sync re-list / side path only)
  -> cursor need not advance (no guaranteed gen_end 2)
```

## Steps

1. Plant only `projects/alpha` for cold.
2. `AddRepoRelPaths=["projects/beta"]` after cold.
3. Override clocks: delta = 5s.
4. Also set DeltaAge=5s so SelectedBudget can be checked when helper exists.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	alpha := filepath.Join(root, "projects", "alpha")
	mkdirAll(t, alpha)
	fakeGitRepo(t, alpha)
	req.Roots = []string{root}
	req.Consume = true
	req.BudgetOnly = false
	req.AddRepoRelPaths = []string{filepath.Join("projects", "beta")}

	t0 := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	req.LastScanEnd = t0
	req.SetLastScanEnd = true
	req.NowAt = t0.Add(5 * time.Second)
	req.SetNow = true
	req.DeltaAge = 5 * time.Second
	return nil
}
```
