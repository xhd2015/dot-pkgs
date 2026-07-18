# Scenario

**Feature**: 10s ≤ delta &lt; 60s selects 500ms sync walk-consume budget

```
# pure tier selection (no FS / Scan)
DeltaAge = 30s
  -> WalkConsumeSyncBudget(30s) == 500ms
  -> also check boundary: 10s → 500ms; 59.999s → 500ms (pragmatic mid-band 30s)
```

## Steps

1. Set `BudgetOnly=true`, `DeltaAge=30s`.
2. Do not require Roots / cold fixtures.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Consume = false
	req.BudgetOnly = true
	req.DeltaAge = 30 * time.Second
	// CacheRoot still set by root Setup; unused for BudgetOnly.
	return nil
}
```
