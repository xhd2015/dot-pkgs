## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `unit-a/still-here` (main).
- Deleted `gone-repo` is **not** listed.
- Index does not list the deleted path.

## Errors

- `err` is nil.

## Side Effects

- Liveness remains binding under budgeted refresh.

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

	stillPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "still-here"))
	wantGitDir := absPath(t, filepath.Join(stillPath, ".git"))
	gonePath := req.RealPath
	if gonePath == "" {
		t.Fatal("req.RealPath empty; Setup must stash deleted repo abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == gonePath {
			t.Fatalf("budgeted warm listed deleted gone-repo at repos[%d]=%q", i, r.Path)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 live repo (still-here), got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	if r.Path != stillPath {
		t.Fatalf("Repos[0].Path = %q, want %q", r.Path, stillPath)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if r.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGitDir)
	}

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if ok {
		for _, e := range idx.Repos {
			if e.Path == gonePath {
				t.Fatalf("index still lists deleted path %s", gonePath)
			}
		}
	}
}
```
