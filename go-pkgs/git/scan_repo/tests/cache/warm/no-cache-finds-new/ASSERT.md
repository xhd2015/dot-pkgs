## Expected

- Scan succeeds with no RootErrors.
- Exactly two repos, path-sorted: `brand-new-repo` then `known-repo`.
- Both are `RepoTypeMain`.
- Proves `NoCache=true` full live walk finds the post-seed plant that warm would miss.

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
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos (full live), got %d", len(resp.Repos))
	}

	brandNew := absPath(t, filepath.Join(req.Roots[0], "brand-new-repo"))
	known := absPath(t, filepath.Join(req.Roots[0], "known-repo"))
	// Path-sorted: brand-new-repo before known-repo
	if resp.Repos[0].Path != brandNew {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, brandNew)
	}
	if resp.Repos[1].Path != known {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, known)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```
