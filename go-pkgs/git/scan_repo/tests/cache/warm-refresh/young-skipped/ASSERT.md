## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `unit-a/known-repo` (main).
- `unit-a/nested/new-repo` is **not** listed (young unit skipped; no rewalk;
  path is not a sibling of any indexed repo).

## Errors

- `err` is nil.

## Side Effects

- YoungAge gate prevents refresh even when WarmRefreshBudget would allow work.
- Soft incompleteness for nested (non-sibling) uncached repos under young units.

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

	knownPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "known-repo"))
	newPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "nested", "new-repo"))

	for i, r := range resp.Repos {
		if r.Path == newPath {
			t.Fatalf("warm refresh listed nested new-repo under young unit at repos[%d]=%q; want omit", i, r.Path)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected exactly 1 repo (known only), got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	if r.Path != knownPath {
		t.Fatalf("Repos[0].Path = %q, want %q", r.Path, knownPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
}
```
