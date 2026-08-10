## Expected

- After `Stop` + `Join`, wall time is well under the full WarmRefreshBudget (3s).
- Join does not return a hard failure solely due to Stop.
- `home/repos.json` still loads (partial writes retained / consistent).

```go
import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {

	start := time.Now()
	budget := 3 * time.Second
	sess, sessErr := scan_repo.ScanSession(context.Background(), scan_repo.Options{
		Roots:             req.Roots,
		CacheRoot:         req.CacheRoot,
		NoCache:           false,
		WarmRefreshMode:   scan_repo.WarmRefreshAsync,
		WarmRefreshBudget: budget,
		YoungAge:          req.YoungAge,
	})
	if sessErr != nil {
		t.Fatal(sessErr)
	}
	if sess.Refresh == nil {
		t.Fatal("expected Refresh handle")
	}
	sess.Stop()
	if joinErr := sess.Join(context.Background()); joinErr != nil {
		t.Fatalf("Join after Stop: %v", joinErr)
	}
	elapsed := time.Since(start)
	// Full budget is 3s; Stop must not force waiting the whole budget.
	if elapsed >= 2*time.Second {
		t.Fatalf("Stop+Join took %s; want well under WarmRefreshBudget=%s", elapsed, budget)
	}

	_, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex after Stop: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected index to remain after Stop (already-scanned retained)")
	}
}
```
