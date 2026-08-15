## Expected

- Exactly two repos, path-sorted: `feature-a` then `main`.
- `feature-a`: `RepoTypeWorktree`, `GitDir` = `main/.git`.
- `main`: `RepoTypeMain`, `GitDir` = `main/.git`.

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
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %v", len(resp.Repos), resp.Repos)
	}
	wantGitDir := absPath(t, filepath.Join(req.Roots[0], "main", ".git"))
	wantFeature := absPath(t, filepath.Join(req.Roots[0], "feature-a"))
	wantMain := absPath(t, filepath.Join(req.Roots[0], "main"))

	if resp.Repos[0].Path != wantFeature {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, wantFeature)
	}
	if resp.Repos[1].Path != wantMain {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, wantMain)
	}
	if resp.Repos[0].RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("feature-a RepoType = %v, want worktree", resp.Repos[0].RepoType)
	}
	if resp.Repos[1].RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("main RepoType = %v, want main", resp.Repos[1].RepoType)
	}
	for _, r := range resp.Repos {
		if r.GitDir != wantGitDir {
			t.Fatalf("%s GitDir = %q, want %q", r.Name, r.GitDir, wantGitDir)
		}
	}
}
```