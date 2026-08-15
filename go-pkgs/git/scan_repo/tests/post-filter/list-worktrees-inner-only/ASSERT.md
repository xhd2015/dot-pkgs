## Expected

- Scan succeeds with no RootErrors.
- A top-level main row exists for `MainPath`.
- That main row’s `Worktrees` includes:
  - `MainPath` with `IsMain=true`
  - `WorktreePath` (`main-wt`) with `IsMain=false`
- Every top-level `Repos` path and every `Worktrees` path is under the
  scan root (`ConsumerPath`).
- **Dual discovery (documented):** FS walk (Option A) may also emit
  `WorktreePath` as its own top-level `RepoTypeWorktree` row when the
  gitlink is under the root. That is allowed. ListWorktrees must not be
  the sole reason for *extra* phantom top-level rows outside FS discovery —
  primary contract is the **inner** `Worktrees` attachment under root.
- Classic TDD: RED if under-root linked worktree is missing from
  `main.Worktrees` after resolve.

## Errors

- `err` is nil.

## Side Effects

- None required beyond git worktree metadata on disk.

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

	root := req.ConsumerPath
	mainPath := req.MainPath
	wtPath := req.WorktreePath
	if root == "" || mainPath == "" || wtPath == "" {
		t.Fatal("Setup must set ConsumerPath, MainPath, WorktreePath")
	}

	// All returned paths under scan root.
	for _, r := range resp.Repos {
		if !pathUnderRoot(root, r.Path) {
			t.Fatalf("top-level repo %q not under scan root %q", r.Path, root)
		}
		for _, wt := range r.Worktrees {
			if wt.Path != "" && !pathUnderRoot(root, wt.Path) {
				t.Fatalf("Worktree %q on %q not under scan root %q", wt.Path, r.Path, root)
			}
		}
	}

	var mainRow *scan_repo.Repo
	for i := range resp.Repos {
		if filepath.Clean(resp.Repos[i].Path) == filepath.Clean(mainPath) {
			mainRow = &resp.Repos[i]
			break
		}
	}
	if mainRow == nil {
		t.Fatalf("expected top-level main %q; repos=%v", mainPath, resp.Repos)
	}
	if mainRow.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("main RepoType = %v, want main", mainRow.RepoType)
	}

	// Primary contract: linked worktree attached on main.Worktrees.
	byPath := map[string]bool{} // path -> IsMain
	for _, wt := range mainRow.Worktrees {
		byPath[filepath.Clean(wt.Path)] = wt.IsMain
	}
	if isMain, ok := byPath[filepath.Clean(mainPath)]; !ok || !isMain {
		t.Fatalf("main.Worktrees missing main path or IsMain false: %v", mainRow.Worktrees)
	}
	if isMain, ok := byPath[filepath.Clean(wtPath)]; !ok || isMain {
		t.Fatalf("main.Worktrees missing under-root linked %q or IsMain true: %v",
			wtPath, mainRow.Worktrees)
	}

	// Dual discovery OK: if worktree also appears top-level, it must be worktree type.
	for _, r := range resp.Repos {
		if filepath.Clean(r.Path) == filepath.Clean(wtPath) && r.RepoType != scan_repo.RepoTypeWorktree {
			t.Fatalf("top-level worktree row RepoType = %v, want worktree", r.RepoType)
		}
	}
}
```
