## Expected

- `MergeBack` returns no error, action is `"noop"`.
- No tmp worktree was created.
- Source worktree still exists.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success for dirty included worktree with no-rm, got: %v", err)
	}
	if resp.Action != "noop" {
		t.Fatalf("expected action 'noop', got %q", resp.Action)
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist after merge-back")
	}
}
```
