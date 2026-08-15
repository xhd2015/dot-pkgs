## Expected

- Scan succeeds with no RootErrors.
- Exactly two repos, path-sorted: `repo-a` then `repo-b`.
- Both `RepoTypeMain` — cold path unlimited despite tiny WarmRefreshBudget.

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
		t.Fatalf("expected 2 repos (cold unlimited), got %d", len(resp.Repos))
	}

	repoA := absPath(t, filepath.Join(req.Roots[0], "repo-a"))
	repoB := absPath(t, filepath.Join(req.Roots[0], "repo-b"))
	if resp.Repos[0].Path != repoA {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, repoA)
	}
	if resp.Repos[1].Path != repoB {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, repoB)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```
