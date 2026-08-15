## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `live-repo` (main).
- `doomed-repo` (`req.DeadPath`) is **not** in `Result.Repos`.
- **Index path (P2):** after Scan, `home/repos.json` loads (`IndexOK`) and does
  **not** list `DeadPath` (liveness applied on the durable index, not only mirror
  marks). Pure `ApplyLiveness` is covered by `cache/repo-index/liveness/`; this
  leaf requires Scan to seed/update the index.

## Errors

- `err` is nil.

## Side Effects

- Stale dead path must not remain in the on-disk index after Scan (omit from
  `repos[]` via ApplyLiveness + save, or equivalent).

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
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	live := req.KnownPath
	dead := req.DeadPath
	if live == "" {
		live = absPath(t, filepath.Join(req.Roots[0], "live-repo"))
	}
	if dead == "" {
		t.Fatal("req.DeadPath empty; Setup must stash doomed abs path")
	}

	for i, r := range resp.Repos {
		if r.Path == dead {
			t.Fatalf("Scan listed dead doomed-repo at repos[%d]=%q", i, r.Path)
		}
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 live repo, got %d: %v", len(resp.Repos), resp.Repos)
	}
	r := resp.Repos[0]
	if r.Path != live {
		t.Fatalf("Path = %q, want %q", r.Path, live)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	wantGit := absPath(t, filepath.Join(live, ".git"))
	if r.GitDir != wantGit {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGit)
	}

	// P2: Scan path must seed/update durable index with liveness applied.
	if !resp.IndexOK {
		t.Fatal("expected IndexOK after Scan; cold must seed home/repos.json and warm keep/update it")
	}
	paths := indexPaths(resp.Index)
	if _, ok := paths[dead]; ok {
		t.Fatalf("index still lists dead path %q after Scan liveness; entries=%v", dead, resp.Index.Repos)
	}
	if _, ok := paths[live]; !ok {
		t.Fatalf("index missing live path %q after Scan; entries=%v", live, resp.Index.Repos)
	}
}
```
