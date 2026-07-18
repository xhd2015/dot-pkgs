## Expected

- Scan succeeds with one discovered main repo.
- `LoadRepoIndex(CacheRoot, "home")` returns ok with an entry for the repo path:
  `repo_type=main`, `git_dir` matching discovery, non-empty `seen_at`.
- `<CacheRoot>/mirror` does not exist.

## Errors

- `err` is nil.

## Side Effects

- Durable index under home universe; dense mirror retired.

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

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected home/repos.json after cold Scan under %s", req.CacheRoot)
	}
	var ie *scan_repo.RepoIndexEntry
	for i := range idx.Repos {
		if idx.Repos[i].Path == repoPath {
			ie = &idx.Repos[i]
			break
		}
	}
	if ie == nil {
		t.Fatalf("index missing %s; entries=%v", repoPath, idx.Repos)
	}
	if ie.RepoType != string(scan_repo.RepoTypeMain) {
		t.Fatalf("index RepoType = %q, want main", ie.RepoType)
	}
	if ie.GitDir != wantGitDir {
		t.Fatalf("index GitDir = %q, want %q", ie.GitDir, wantGitDir)
	}
	if ie.SeenAt == "" {
		t.Fatal("SeenAt empty, want non-empty RFC3339")
	}
	if _, parseErr := time.Parse(time.RFC3339, ie.SeenAt); parseErr != nil {
		if _, parseErr2 := time.Parse(time.RFC3339Nano, ie.SeenAt); parseErr2 != nil {
			t.Fatalf("SeenAt %q not RFC3339: %v", ie.SeenAt, parseErr)
		}
	}

	assertNoMirrorDir(t, req.CacheRoot)
}
```
