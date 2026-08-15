## Expected

- `MergeBack` returns an error containing "worktree is not clean".
- Source worktree still exists.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for dirty worktree with --rm")
	}
	if !strings.Contains(err.Error(), "has uncommitted changes") {
		t.Fatalf("expected 'has uncommitted changes' error, got: %v", err)
	}
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist after error")
	}
}
```
