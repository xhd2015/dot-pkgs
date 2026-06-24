## Expected

- Two scan rows: `feature-a` (worktree) and `main` (main), path-sorted.
- Only `main` row has `Worktrees` with two entries.
- Main worktree has `IsMain=true`; linked has `IsMain=false`.

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
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %v", len(resp.Repos), resp.Repos)
	}
	wantMain := absPath(t, filepath.Join(req.Roots[0], "main"))
	wantFeature := absPath(t, filepath.Join(req.Roots[0], "feature-a"))

	var mainRow, wtRow *scan_repo.Repo
	for i := range resp.Repos {
		switch resp.Repos[i].Name {
		case "main":
			mainRow = &resp.Repos[i]
		case "feature-a":
			wtRow = &resp.Repos[i]
		}
	}
	if mainRow == nil || wtRow == nil {
		t.Fatalf("expected main and feature-a rows, got %v", resp.Repos)
	}
	if wtRow.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("feature-a RepoType = %v, want worktree", wtRow.RepoType)
	}
	if len(wtRow.Worktrees) != 0 {
		t.Fatalf("worktree row should have empty Worktrees, got %v", wtRow.Worktrees)
	}
	if len(mainRow.Worktrees) != 2 {
		t.Fatalf("main row should have 2 worktrees, got %v", mainRow.Worktrees)
	}
	byPath := map[string]bool{}
	for _, wt := range mainRow.Worktrees {
		byPath[wt.Path] = wt.IsMain
	}
	if isMain, ok := byPath[wantMain]; !ok || !isMain {
		t.Fatalf("main worktree missing or IsMain false: %v", mainRow.Worktrees)
	}
	if isMain, ok := byPath[wantFeature]; !ok || isMain {
		t.Fatalf("feature-a worktree missing or IsMain true: %v", mainRow.Worktrees)
	}
}
```