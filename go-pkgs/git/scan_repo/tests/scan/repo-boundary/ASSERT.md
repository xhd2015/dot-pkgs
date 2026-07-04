## Expected

- Two repos: `outer` and nested `inner`, path-sorted ascending.
- Both are `RepoTypeMain`.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos (outer + inner), got %d: %v", len(resp.Repos), resp.Repos)
	}
	wantOuter := absPath(t, filepath.Join(req.Roots[0], "outer"))
	wantInner := absPath(t, filepath.Join(req.Roots[0], "outer", "inner"))
	if resp.Repos[0].Path != wantOuter {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, wantOuter)
	}
	if resp.Repos[1].Path != wantInner {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, wantInner)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```