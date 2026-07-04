## Expected

- Two repos discovered when scanning from the worktree root.
- Worktree root row: `RepoTypeWorktree`.
- `vendor/nested` row: `RepoTypeMain`.

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
		t.Fatalf("expected 2 repos (worktree root + nested main repo), got %d: %v", len(resp.Repos), resp.Repos)
	}

	scanRoot := absPath(t, req.Roots[0])
	wantNested := absPath(t, filepath.Join(scanRoot, "vendor", "nested"))

	var wtRow, nestedRow *scan_repo.Repo
	for i := range resp.Repos {
		switch resp.Repos[i].Path {
		case scanRoot:
			wtRow = &resp.Repos[i]
		case wantNested:
			nestedRow = &resp.Repos[i]
		}
	}
	if wtRow == nil {
		t.Fatalf("missing worktree root row at %q, got %v", scanRoot, resp.Repos)
	}
	if wtRow.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("worktree root RepoType = %v, want worktree", wtRow.RepoType)
	}
	if nestedRow == nil {
		t.Fatalf("missing nested main repo at %q, got %v", wantNested, resp.Repos)
	}
	if nestedRow.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("nested RepoType = %v, want main", nestedRow.RepoType)
	}
}
```