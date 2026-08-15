## Expected

- Discovery includes `feature-a` as `RepoTypeWorktree` with `GitDir` → `main/.git`.
- Index entry for `feature-a`: `repo_type="worktree"`, same `git_dir`.
- Index entry for `main`: `repo_type="main"`.
- No `mirror/` under CacheRoot.

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

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected home/repos.json after cold Scan")
	}
	byPath := map[string]scan_repo.RepoIndexEntry{}
	for _, e := range idx.Repos {
		byPath[e.Path] = e
	}
	wtE, ok := byPath[wtPath]
	if !ok {
		t.Fatalf("index missing worktree %s", wtPath)
	}
	if wtE.RepoType != string(scan_repo.RepoTypeWorktree) {
		t.Fatalf("worktree index RepoType=%q, want worktree", wtE.RepoType)
	}
	if wtE.GitDir != wantGitDir {
		t.Fatalf("worktree index GitDir=%q, want %q", wtE.GitDir, wantGitDir)
	}
	mainE, ok := byPath[mainPath]
	if !ok {
		t.Fatalf("index missing main %s", mainPath)
	}
	if mainE.RepoType != string(scan_repo.RepoTypeMain) {
		t.Fatalf("main index RepoType=%q, want main", mainE.RepoType)
	}
	assertNoMirrorDir(t, req.CacheRoot)
}
```
