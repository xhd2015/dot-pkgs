## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `still-here` (main), correct Path/GitDir.
- `gone-repo` is **not** listed in `resp.Repos`.
- Mirror for the deleted path (`req.RealPath`): either Load returns `ok=false`,
  or `ok=true` with `is_repo=false` (liveness cleared the mark).

## Errors

- `err` is nil.

## Side Effects

- Stale `is_repo=true` for a deleted path must not survive warm Scan.

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

	stillPath := absPath(t, filepath.Join(req.Roots[0], "still-here"))
	wantGitDir := absPath(t, filepath.Join(stillPath, ".git"))
	gonePath := req.RealPath
	if gonePath == "" {
		t.Fatal("req.RealPath empty; Setup must stash deleted repo abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == gonePath {
			t.Fatalf("warm Scan listed deleted gone-repo at repos[%d]=%q", i, r.Path)
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

	// Liveness must clear the stale mirror mark for the deleted path.
	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, gonePath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(gone): %v", loadErr)
	}
	if ok && entry.IsRepo {
		t.Fatalf("deleted path still is_repo=true in cache (want absent or is_repo=false); entry=%+v", entry)
	}
}
```
