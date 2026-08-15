# Scenario

**Feature**: delta ≥ 60s selects 1s sync budget; consume can seal gen_end 2

```
# tier + light behavioral smoke
WalkConsumeSyncBudget(60s) == 1s
WalkConsumeSyncBudget(2m) == 1s
  + cold then second Scan with delta=2m
  -> gen_end 2 present (1s budget enough for tiny fixture)
```

## Steps

1. Plant small workspace (alpha).
2. Keep grouping default clocks (delta 2m) and `Consume=true`.
3. Assert both pure 1s tier and gen_end 2 seal under that budget.

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	alpha := filepath.Join(root, "projects", "alpha")
	mkdirAll(t, alpha)
	fakeGitRepo(t, alpha)
	req.Roots = []string{root}
	req.Consume = true
	req.BudgetOnly = false

	// Explicit ≥60s (grouping already sets 2m; restate for clarity).
	t0 := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	req.LastScanEnd = t0
	req.SetLastScanEnd = true
	req.NowAt = t0.Add(2 * time.Minute)
	req.SetNow = true
	req.DeltaAge = 2 * time.Minute
	return nil
}
```
