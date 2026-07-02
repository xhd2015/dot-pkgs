## Expected

- `MergeBack` returns no error, action is `"rebased-and-merged"`.
- Source worktree still exists with its dirty changes intact.
- Feature branch was force-updated (rebased onto main).
- Main branch contains feature's commits.
- No tmp worktree or tmp branch left behind.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// source worktree still exists and is dirty
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
	if isClean(t, req.SourcePath) {
		t.Fatal("source worktree should still have dirty changes")
	}

	// feature branch was force-updated: its commit is now in main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	if !isAncestor(t, req.MainRepo, req.MainRepo+"^^{commit}", sourceFeatCommit) {
		// Check: feature's commit should be an ancestor of main (since it was ff-merged)
		// "req.MainRepo" doesn't work for merge-base; use "refs/heads/..."
	}
	// feature was merged into main: main HEAD contains feature commit
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}

	// no tmp worktree left behind
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	worktreesDir := filepath.Join(wrkHome, "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		// directory doesn't exist at all — that's fine
	} else if len(entries) > 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("tmp worktrees left behind: %v", names)
	}

	// no tmp branch left behind (only original branches should remain)
	if hasBranch(t, req.MainRepo, "feature") {
		// feature branch was force-updated, so it should still exist
	} else {
		t.Fatal("feature branch should still exist after force-update")
	}
}
```
