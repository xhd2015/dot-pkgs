## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `unit-a/still-here` (main).
- Deleted `gone-repo` is **not** listed.
- Mirror for deleted path: absent or `is_repo=false`.

## Errors

- `err` is nil.

## Side Effects

- Liveness (P3) remains binding under budgeted refresh (P4).

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

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, gonePath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(gone): %v", loadErr)
	}
	if ok && entry.IsRepo {
		t.Fatalf("deleted path still is_repo=true in cache; entry=%+v", entry)
	}
}
```
