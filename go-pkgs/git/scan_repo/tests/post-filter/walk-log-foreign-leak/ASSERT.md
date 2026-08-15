## Expected

- Scan succeeds with no RootErrors.
- Every `Result.Repos` path is under abs scan root `ConsumerPath`.
- Foreign `agent-pro` (`ForeignPath`) **must not** appear in `Result.Repos`.
- Consumer itself is present (warm-eligible indexed live root).
- `ListWorktrees` is false — no expand; contract is top-level return filter
  after walk-log consume.
- Classic TDD: **RED** while `consumeWalkLog` merges foreign checkouts
  without post-process base-path filter.

## Errors

- `err` is nil.

## Side Effects

- Walk log / index may grow during consume; not asserted — Result path set
  is the contract.

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
	if req.ListWorktrees {
		t.Fatal("ListWorktrees must be false for this leaf")
	}

	consumer := req.ConsumerPath
	foreign := req.ForeignPath
	if consumer == "" || foreign == "" {
		t.Fatal("Setup must set ConsumerPath and ForeignPath")
	}
	if len(req.Roots) != 1 || absPath(t, req.Roots[0]) != consumer {
		t.Fatalf("Roots must be single consumer %q, got %v", consumer, req.Roots)
	}

	got := resultPaths(resp.Repos)
	if _, ok := got[filepath.Clean(foreign)]; ok {
		t.Fatalf("Result must omit foreign agent-pro %q outside consumer %q; repos=%v",
			foreign, consumer, resp.Repos)
	}
	// No path may escape the scan root (foreign leak or any other checkout).
	for _, r := range resp.Repos {
		if !pathUnderRoot(consumer, r.Path) {
			t.Fatalf("repo path %q not under scan root %q (foreign leak)", r.Path, consumer)
		}
		// Nested Worktrees (if any) must also stay under root when present.
		for _, wt := range r.Worktrees {
			if wt.Path != "" && !pathUnderRoot(consumer, wt.Path) {
				t.Fatalf("Worktree path %q not under scan root %q", wt.Path, consumer)
			}
		}
	}
	if _, ok := got[filepath.Clean(consumer)]; !ok {
		t.Fatalf("Result missing cold-indexed consumer %q; repos=%v", consumer, resp.Repos)
	}
	for i, r := range resp.Repos {
		if filepath.Clean(r.Path) == filepath.Clean(consumer) && r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("consumer repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```
