## Expected

- `MergeBack` returns no error, action is `"rebased-and-merged"`.
- The tmp worktree with suffix -1 was cleaned up (it should not exist after merge).
- The collision dir we pre-created still exists.
- Feature branch was updated.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// source worktree still exists
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	worktreesDir := filepath.Join(wrkHome, "worktrees")

	date := time.Now().Format("2006-01-02")
	collisionDir := filepath.Join(worktreesDir, "main-feature-"+date+"-tmp-rebase")
	suffixDir := filepath.Join(worktreesDir, "main-feature-"+date+"-tmp-rebase-1")

	// collision dir (pre-created) still exists
	if !hasDir(collisionDir) {
		t.Fatal("pre-created collision dir should still exist")
	}

	// suffix dir was cleaned up
	if hasDir(suffixDir) {
		t.Fatal("tmp worktree with suffix should have been cleaned up")
	}

	// feature was merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}
}
```
