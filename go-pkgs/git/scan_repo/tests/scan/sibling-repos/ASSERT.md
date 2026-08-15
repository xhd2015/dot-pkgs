## Expected

- Exactly two repos in `resp.Repos`.
- Sorted by `Path` ascending: `alpha` before `zebra`.
- Both are `RepoTypeMain`.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(resp.Repos))
	}
	alpha := absPath(t, filepath.Join(req.Roots[0], "alpha"))
	zebra := absPath(t, filepath.Join(req.Roots[0], "zebra"))
	if resp.Repos[0].Path != alpha {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, alpha)
	}
	if resp.Repos[1].Path != zebra {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, zebra)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```