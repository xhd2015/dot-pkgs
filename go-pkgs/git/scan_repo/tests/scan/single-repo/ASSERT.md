## Expected

- Exactly one repo in `resp.Repos`.
- `RepoType` is `RepoTypeMain`.
- `Name` is `"my-repo"`.
- `GitDir` is `<repo>/.git` (absolute, cleaned).
- `Remotes` and `Worktrees` are empty/nil.

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
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	wantPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	wantGitDir := absPath(t, filepath.Join(wantPath, ".git"))
	if r.Path != wantPath {
		t.Fatalf("Path = %q, want %q", r.Path, wantPath)
	}
	if r.Name != "my-repo" {
		t.Fatalf("Name = %q, want my-repo", r.Name)
	}
	if r.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGitDir)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if len(r.Remotes) != 0 {
		t.Fatalf("expected empty Remotes, got %v", r.Remotes)
	}
	if len(r.Worktrees) != 0 {
		t.Fatalf("expected empty Worktrees, got %v", r.Worktrees)
	}
}
```