## Expected

- `MergeBack` returns an error containing "rebase conflict".
- Source worktree still exists with dirty changes.
- Feature branch was NOT force-updated (still at original commit).
- No tmp worktree or tmp branch left behind.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error from rebase conflict")
	}
	if !strings.Contains(err.Error(), "rebase conflict") {
		t.Fatalf("expected 'rebase conflict' error, got: %v", err)
	}

	// source worktree still exists
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist after conflict")
	}

	// no tmp worktree left behind
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	worktreesDir := filepath.Join(wrkHome, "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	} else if len(entries) > 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("tmp worktrees left behind: %v", names)
	}

	// feature branch still exists (was not deleted during abort)
	if !hasBranch(t, req.MainRepo, "feature") {
		t.Fatal("feature branch should still exist after abort")
	}
}
```
