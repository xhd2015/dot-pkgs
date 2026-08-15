## Expected

- Scan succeeds with no RootErrors.
- At least the main checkout is present as a top-level row under the scan
  root (`MainPath` / `ConsumerPath`).
- On the main row (when present), `Worktrees` must **not** include the outer
  linked path `WorktreePath` / `ForeignPath`.
- Every `Worktrees` path on every returned repo is under the scan root
  (typically only the main itself with `IsMain=true`, if Worktrees is
  non-empty after strip).
- No top-level `Repos` path outside the scan root.
- Classic TDD: **RED** while git porcelain outer worktrees remain attached
  without post-resolve base-path filter.

## Errors

- `err` is nil.

## Side Effects

- Outer worktree still exists on disk; filter is return-value only.

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
	if !req.ListWorktrees {
		t.Fatal("ListWorktrees must be true for this leaf")
	}

	mainPath := req.MainPath
	outer := req.WorktreePath
	if mainPath == "" || outer == "" {
		t.Fatal("Setup must set MainPath and WorktreePath")
	}
	if pathUnderRoot(mainPath, outer) {
		t.Fatalf("fixture error: outer %q must not be under scan root %q", outer, mainPath)
	}

	// No top-level escape.
	for _, r := range resp.Repos {
		if !pathUnderRoot(mainPath, r.Path) {
			t.Fatalf("top-level repo %q not under scan root %q", r.Path, mainPath)
		}
	}

	got := resultPaths(resp.Repos)
	if _, ok := got[filepath.Clean(mainPath)]; !ok {
		t.Fatalf("Result missing main %q; repos=%v", mainPath, resp.Repos)
	}
	// Outer must never be a top-level row either (outside base).
	if _, ok := got[filepath.Clean(outer)]; ok {
		t.Fatalf("Result must omit outer worktree top-level %q; repos=%v", outer, resp.Repos)
	}

	// Core: strip outer from Worktrees after ListWorktrees resolve.
	for _, r := range resp.Repos {
		for _, wt := range r.Worktrees {
			clean := filepath.Clean(wt.Path)
			if clean == filepath.Clean(outer) {
				t.Fatalf("Worktrees on %q still lists outer path %q (must strip outside base); Worktrees=%v",
					r.Path, outer, r.Worktrees)
			}
			if wt.Path != "" && !pathUnderRoot(mainPath, wt.Path) {
				t.Fatalf("Worktree path %q on %q not under scan root %q", wt.Path, r.Path, mainPath)
			}
		}
	}

	// If main row has Worktrees non-empty after filter, main itself may remain as IsMain.
	for i := range resp.Repos {
		r := &resp.Repos[i]
		if filepath.Clean(r.Path) != filepath.Clean(mainPath) {
			continue
		}
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("main RepoType = %v, want main", r.RepoType)
		}
		for _, wt := range r.Worktrees {
			if filepath.Clean(wt.Path) == filepath.Clean(mainPath) && !wt.IsMain {
				t.Fatalf("main worktree entry IsMain=false: %v", r.Worktrees)
			}
		}
	}
}
```
