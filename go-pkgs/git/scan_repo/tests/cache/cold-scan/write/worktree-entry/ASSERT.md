## Expected

- Discovery includes `feature-a` as `RepoTypeWorktree` with `GitDir` → `main/.git`.
- Mirror entry for `feature-a`: `is_repo=true`, `repo_type="worktree"`, same `git_dir`.
- Main checkout also has an `is_repo` entry with `repo_type="main"`.

## Errors

- `err` is nil.

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
	wtPath := absPath(t, filepath.Join(req.Roots[0], "feature-a"))
	mainPath := absPath(t, filepath.Join(req.Roots[0], "main"))
	wantGitDir := absPath(t, filepath.Join(mainPath, ".git"))

	var wt *scan_repo.Repo
	for i := range resp.Repos {
		if resp.Repos[i].Path == wtPath {
			wt = &resp.Repos[i]
			break
		}
	}
	if wt == nil {
		t.Fatalf("expected worktree row %s, got %v", wtPath, resp.Repos)
	}
	if wt.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("RepoType = %v, want worktree", wt.RepoType)
	}
	if wt.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", wt.GitDir, wantGitDir)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, wtPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(worktree): %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected cache entry for worktree %s", wtPath)
	}
	if !entry.IsRepo {
		t.Fatal("worktree IsRepo=false, want true")
	}
	if entry.RepoType != string(scan_repo.RepoTypeWorktree) {
		t.Fatalf("worktree RepoType=%q, want worktree", entry.RepoType)
	}
	if entry.GitDir != wantGitDir {
		t.Fatalf("worktree cache GitDir=%q, want %q", entry.GitDir, wantGitDir)
	}
	if !entry.ScanComplete {
		t.Fatal("worktree ScanComplete=false")
	}

	mainEntry, mainOK, mainErr := scan_repo.LoadCacheEntry(req.CacheRoot, mainPath)
	if mainErr != nil {
		t.Fatalf("LoadCacheEntry(main): %v", mainErr)
	}
	if !mainOK {
		t.Fatalf("expected cache entry for main %s", mainPath)
	}
	if !mainEntry.IsRepo || mainEntry.RepoType != string(scan_repo.RepoTypeMain) {
		t.Fatalf("main entry IsRepo=%v RepoType=%q, want true/main", mainEntry.IsRepo, mainEntry.RepoType)
	}
}
```
