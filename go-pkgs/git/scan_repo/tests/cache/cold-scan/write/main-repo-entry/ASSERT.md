## Expected

- Scan succeeds with at least one discovered main repo.
- After Scan, `LoadCacheEntry(CacheRoot, repoPath)` returns `ok=true`.
- Entry fields for the repo directory:
  - `version=1`
  - `is_repo=true`
  - `repo_type="main"`
  - `git_dir` equals discovered `Repo.GitDir` (absolute)
  - `refreshed_at` non-empty RFC3339
  - `scan_complete=true`

## Errors

- `err` is nil.

## Side Effects

- Mirror `entry.json` exists for the repo path under `CacheRoot/mirror/...`.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	wantGitDir := absPath(t, filepath.Join(repoPath, ".git"))

	var found *scan_repo.Repo
	for i := range resp.Repos {
		if resp.Repos[i].Path == repoPath {
			found = &resp.Repos[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected discovery of %s, got %v", repoPath, resp.Repos)
	}
	if found.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", found.RepoType)
	}
	if found.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", found.GitDir, wantGitDir)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, repoPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry: %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected cache entry for repo %s under %s", repoPath, req.CacheRoot)
	}
	if entry.Version != 1 {
		t.Fatalf("Version = %d, want 1", entry.Version)
	}
	if !entry.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if entry.RepoType != string(scan_repo.RepoTypeMain) {
		t.Fatalf("RepoType = %q, want %q", entry.RepoType, scan_repo.RepoTypeMain)
	}
	if entry.GitDir != wantGitDir {
		t.Fatalf("cache GitDir = %q, want %q", entry.GitDir, wantGitDir)
	}
	if !entry.ScanComplete {
		t.Fatal("ScanComplete = false, want true")
	}
	if entry.RefreshedAt == "" {
		t.Fatal("RefreshedAt empty, want non-empty RFC3339")
	}
	if _, parseErr := time.Parse(time.RFC3339, entry.RefreshedAt); parseErr != nil {
		// also accept RFC3339Nano
		if _, parseErr2 := time.Parse(time.RFC3339Nano, entry.RefreshedAt); parseErr2 != nil {
			t.Fatalf("RefreshedAt %q not RFC3339: %v", entry.RefreshedAt, parseErr)
		}
	}
}
```
