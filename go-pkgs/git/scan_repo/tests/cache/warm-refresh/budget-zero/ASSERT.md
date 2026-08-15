## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `unit-a/known-repo` (warm serve still works).
- `unit-a/nested/new-repo` is **not** listed (no unit rewalk under negative budget;
  path is not a sibling of any indexed repo).

## Errors

- `err` is nil.

## Side Effects

- Documents budget gate: negative WarmRefreshBudget equals pure P3 warm serve
  (index + sibling probe only; no nested discovery under aged units).

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
			t.Fatalf("listed nested new-repo at repos[%d] with WarmRefreshBudget=-1; want omit (no unit rewalk)", i)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo (known only), got %d", len(resp.Repos))
	}
	if resp.Repos[0].Path != knownPath {
		t.Fatalf("Repos[0].Path = %q, want %q", resp.Repos[0].Path, knownPath)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", resp.Repos[0].RepoType)
	}
}
```
