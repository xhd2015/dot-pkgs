## Expected

- Scan completes with `err == nil` (budget expiry must not hard-fail via parent
  `ctx` / SIGINT path).
- No `RootErrors` for the scan root: unit budget cancel is soft — return partial
  merge of warm-served (and any mid-walk discoveries), not a root failure.
- `unit-a/known-repo` is still in `Result.Repos` (warm serve / partial merge).
- Wall time for this Scan (`resp.Elapsed`) is **under 2s** — much less than the
  multi-second unbounded rewalk of the padded unit (proves budget is enforced
  **inside** the unit walk, not only between units).

## Errors

- `err` is nil.

## Side Effects

- Documents P4 mid-unit budget: child context with deadline (or equivalent) for
  the unit rewalk only; parent Scan context stays live so warm results survive.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// maxScanWall is a generous upper bound for budgeted warm Scan (100ms budget +
// serve + overhead). Unbounded walkRoot of the pad tree is multi-second.
const maxScanWall = 2 * time.Second

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Scan hard-failed under tiny WarmRefreshBudget: %v (budget cancel must not abort via parent ctx)", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors after budget-capped unit walk (soft partial merge), got %v", resp.RootErrors)
	}

	knownPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "known-repo"))
	foundKnown := false
	for _, r := range resp.Repos {
		if r.Path == knownPath {
			foundKnown = true
			if r.RepoType != scan_repo.RepoTypeMain {
				t.Fatalf("known-repo RepoType = %v, want main", r.RepoType)
			}
			break
		}
	}
	if !foundKnown {
		t.Fatalf("missing warm-served known-repo %q in Result (repos=%v); budget mid-cancel must keep prior serve", knownPath, pathsOf(resp.Repos))
	}

	if resp.Elapsed <= 0 {
		t.Fatal("resp.Elapsed unset; harness Run must time Scan")
	}
	if resp.Elapsed >= maxScanWall {
		t.Fatalf("Scan wall time %v >= %v; WarmRefreshBudget=%v must cap mid-unit walkRoot (today budget is only checked between units)",
			resp.Elapsed, maxScanWall, req.WarmRefreshBudget)
	}
}

func pathsOf(repos []scan_repo.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Path
	}
	return out
}
```
