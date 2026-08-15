## Expected

- Two repos discovered when scanning from the worktree root.
- `consumer-wt` row: `RepoTypeWorktree`.
- `mydep` row under `external/mydep`: `RepoTypeWorktree`.
- Results path-sorted ascending.

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
		t.Fatalf("expected 2 repos (worktree root + nested external linked wt), got %d: %v", len(resp.Repos), resp.Repos)
	}

	scanRoot := absPath(t, req.Roots[0])
	wantConsumerWt := scanRoot
	wantExternalWt := absPath(t, filepath.Join(scanRoot, "external", "mydep"))

	byName := map[string]scan_repo.Repo{}
	for _, r := range resp.Repos {
		byName[r.Name] = r
	}

	consumerRow, ok := byName["consumer-wt"]
	if !ok {
		// scan root basename may differ; match by Path
		for _, r := range resp.Repos {
			if r.Path == wantConsumerWt {
				consumerRow = r
				ok = true
				break
			}
		}
	}
	if !ok {
		t.Fatalf("missing worktree root row; want Path=%q, got %v", wantConsumerWt, resp.Repos)
	}
	if consumerRow.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("worktree root RepoType = %v, want worktree", consumerRow.RepoType)
	}

	externalRow, ok := byName["mydep"]
	if !ok {
		t.Fatalf("missing nested external linked wt row; want Path=%q, got %v", wantExternalWt, resp.Repos)
	}
	if externalRow.Path != wantExternalWt {
		t.Fatalf("external Path = %q, want %q", externalRow.Path, wantExternalWt)
	}
	if externalRow.RepoType != scan_repo.RepoTypeWorktree {
		t.Fatalf("external RepoType = %v, want worktree", externalRow.RepoType)
	}

	if resp.Repos[0].Path > resp.Repos[1].Path {
		t.Fatalf("repos not path-sorted: %v", resp.Repos)
	}
}
```