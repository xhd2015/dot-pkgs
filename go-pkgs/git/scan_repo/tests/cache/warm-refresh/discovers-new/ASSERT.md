## Expected

- Scan succeeds with no RootErrors.
- Exactly two repos, path-sorted: `known-repo` then `new-repo` under `unit-a`.
- Both are `RepoTypeMain` with correct Path/GitDir.
- Mirror has `is_repo=true` for `new-repo` after budgeted refresh.

## Errors

- `err` is nil.

## Side Effects

- Refresh rewalk of aged unit writes the new repo into the mirror cache.

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

	knownPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "known-repo"))
	newPath := absPath(t, filepath.Join(req.Roots[0], "unit-a", "new-repo"))
	wantKnownGit := absPath(t, filepath.Join(knownPath, ".git"))
	wantNewGit := absPath(t, filepath.Join(newPath, ".git"))

	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos (known + refreshed new), got %d: %v", len(resp.Repos), pathsOf(resp.Repos))
	}
	// path-sorted: known-repo before new-repo
	if resp.Repos[0].Path != knownPath {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, knownPath)
	}
	if resp.Repos[1].Path != newPath {
		t.Fatalf("repos[1].Path = %q, want %q (budgeted refresh must discover new under aged unit)", resp.Repos[1].Path, newPath)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeMain || resp.Repos[0].GitDir != wantKnownGit {
		t.Fatalf("known-repo shape: type=%v gitDir=%q", resp.Repos[0].RepoType, resp.Repos[0].GitDir)
	}
	if resp.Repos[1].RepoType != scan_repo.RepoTypeMain || resp.Repos[1].GitDir != wantNewGit {
		t.Fatalf("new-repo shape: type=%v gitDir=%q", resp.Repos[1].RepoType, resp.Repos[1].GitDir)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, newPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(new-repo): %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected mirror entry for new-repo after refresh: %s", newPath)
	}
	if !entry.IsRepo {
		t.Fatalf("new-repo cache IsRepo=false, want true after refresh")
	}
}

func pathsOf(repos []scan_repo.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Path
	}
	return out
}
```
