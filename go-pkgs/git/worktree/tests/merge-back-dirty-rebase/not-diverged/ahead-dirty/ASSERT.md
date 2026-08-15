## Expected

- `MergeBack` returns no error, action is `"merged"`.
- Source worktree still exists with its dirty changes.
- Feature branch was merged into main.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success for dirty ahead worktree with no-rm, got: %v", err)
	}
	if resp.Action != "merged" {
		t.Fatalf("expected action 'merged', got %q", resp.Action)
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}
	// feature was merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}
}
```
