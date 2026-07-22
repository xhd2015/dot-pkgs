## Expected

- Scan succeeds with no RootErrors.
- `ListWorktrees` is false — every returned repo has empty/nil `Worktrees`
  (no expand).
- Every `Result.Repos` path is under abs scan root `A` (`ConsumerPath`).
- Neighbor sibling `B` (`SiblingPath`) **must not** appear.
- At least `A` itself is present as `RepoTypeMain`.
- Classic TDD: RED while sibling probe / other paths merge `B` without
  under-root drop, or while flag-off still fills `Worktrees`.

## Errors

- `err` is nil.

## Side Effects

- Index save-back may update home universe; Result path set + empty
  Worktrees are the contract.

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
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}
	if req.ListWorktrees {
		t.Fatal("ListWorktrees must be false for this leaf")
	}

	a := req.ConsumerPath
	b := req.SiblingPath
	if a == "" || b == "" {
		t.Fatal("Setup must set ConsumerPath (A) and SiblingPath (B)")
	}
	if len(req.Roots) != 1 || absPath(t, req.Roots[0]) != a {
		t.Fatalf("Roots must be single child A %q, got %v", a, req.Roots)
	}

	got := resultPaths(resp.Repos)
	if _, ok := got[filepath.Clean(b)]; ok {
		t.Fatalf("Result must omit sibling B %q outside scan root A %q; repos=%v",
			b, a, resp.Repos)
	}
	for _, r := range resp.Repos {
		if !pathUnderRoot(a, r.Path) {
			t.Fatalf("repo path %q not under scan root %q", r.Path, a)
		}
		if len(r.Worktrees) != 0 {
			t.Fatalf("%s: expected empty Worktrees when ListWorktrees=false, got %v",
				r.Path, r.Worktrees)
		}
	}
	if _, ok := got[filepath.Clean(a)]; !ok {
		t.Fatalf("Result missing cold-indexed A %q; repos=%v", a, resp.Repos)
	}
	for i, r := range resp.Repos {
		if filepath.Clean(r.Path) == filepath.Clean(a) && r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("A repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```
