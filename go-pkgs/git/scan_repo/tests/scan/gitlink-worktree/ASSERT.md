## Expected

- Worktree row exists for `feature-a` with `RepoTypeWorktree`.
- `GitDir` resolves to `main/.git` (absolute).

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
	var wt *scan_repo.Repo
	for i := range resp.Repos {
		if resp.Repos[i].Name == "feature-a" {
			wt = &resp.Repos[i]
			break
		}
	}
	if wt == nil {
		t.Fatalf("expected feature-a worktree row, got %v", resp.Repos)
	}
	wantPath := absPath(t, filepath.Join(req.Roots[0], "feature-a"))
	wantGitDir := absPath(t, filepath.Join(req.Roots[0], "main", ".git"))
	if wt.Path != wantPath {
		t.Fatalf("Path = %q, want %q", wt.Path, wantPath)
	}
	if wt.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", wt.GitDir, wantGitDir)
	}
	if wt.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("RepoType = %v, want worktree", wt.RepoType)
	}
}
```