## Expected

- `WalkConsumeSyncBudget(5s)` / `resp.SelectedBudget` is **0** when the helper
  is available (`SelectedBudgetOK`).
- Second Scan `Result` includes main `projects/alpha` (warm serve / prior
  index) but **does not** include `projects/beta` planted only after cold —
  proving zero **sync** walk-consume discover under delta &lt; 10s.
- Document: product may still schedule side/best-effort work asynchronously;
  this leaf only forbids sync discovery of beta within this Scan.

## Errors

- `err` is nil (Scan must not hard-fail merely because budget is 0).

## Side Effects

- Optional: cursor may remain at cold EOF when no sync consume runs.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// Tier table: <10s → 0.
	wantBudget := time.Duration(0)
	got := scan_repo.WalkConsumeSyncBudget(5 * time.Second)
	if got != wantBudget {
		t.Fatalf("WalkConsumeSyncBudget(5s) = %v, want %v", got, wantBudget)
	}
	if resp.SelectedBudgetOK && resp.SelectedBudget != wantBudget {
		t.Fatalf("SelectedBudget = %v, want %v", resp.SelectedBudget, wantBudget)
	}

	alpha := absPath(t, filepath.Join(req.Roots[0], "projects", "alpha"))
	beta := absPath(t, filepath.Join(req.Roots[0], "projects", "beta"))

	foundAlpha, foundBeta := false, false
	for _, r := range resp.Repos {
		p := absPath(t, r.Path)
		if p == alpha && r.RepoType == scan_repo.RepoTypeMain {
			foundAlpha = true
		}
		if p == beta {
			foundBeta = true
		}
	}
	if !foundAlpha {
		t.Fatalf("expected warm/served alpha %q; repos=%v", alpha, resp.Repos)
	}
	if foundBeta {
		t.Fatalf("delta<10s must not sync-discover new beta %q via walk-consume; repos=%v",
			beta, resp.Repos)
	}
}
```
