## Expected

- `MergeBack` returns no error, action is `"rebased-and-merged"`.
- Source worktree exists and is clean (rebased in place).
- Feature branch was merged into main.
- No tmp worktree was created at `~/.wrk/worktrees/`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// source worktree still exists and is clean (rebased in place)
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
	if !isClean(t, req.SourcePath) {
		t.Fatal("source worktree should be clean after direct rebase")
	}

	// feature was merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}

	// no tmp worktree was created
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	worktreesDir := filepath.Join(wrkHome, "worktrees")
	entries, errDir := os.ReadDir(worktreesDir)
	if errDir != nil {
		if !os.IsNotExist(errDir) {
			t.Fatal(errDir)
		}
	} else if len(entries) > 0 {
		t.Fatalf("no tmp worktree should exist for clean rebase, got: %d entries", len(entries))
	}
}
```
