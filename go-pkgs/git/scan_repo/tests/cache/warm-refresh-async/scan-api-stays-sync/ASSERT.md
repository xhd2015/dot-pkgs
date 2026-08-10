## Expected

- `Scan` with `WarmRefreshMode: WarmRefreshAsync` still behaves as sync:
  Result includes known-repo and new-repo after return (no Join needed).

```go
import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {

	result, scanErr := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:             req.Roots,
		CacheRoot:         req.CacheRoot,
		NoCache:           false,
		WarmRefreshMode:   scan_repo.WarmRefreshAsync, // must be forced Sync by Scan
		WarmRefreshBudget: 0,                          // product default 1s
		YoungAge:          req.YoungAge,
	})
	if scanErr != nil {
		t.Fatal(scanErr)
	}
	knownPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "known-repo"))
	newPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "nested", "new-repo"))
	if len(result.Repos) != 2 {
		t.Fatalf("Scan must stay sync (2 repos), got %d", len(result.Repos))
	}
	if result.Repos[0].Path != knownPath || result.Repos[1].Path != newPath {
		t.Fatalf("paths: got %q %q", result.Repos[0].Path, result.Repos[1].Path)
	}
}
```
