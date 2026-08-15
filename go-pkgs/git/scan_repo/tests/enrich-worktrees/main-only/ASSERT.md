## Expected

- One repo row (`main`).
- `Worktrees` has exactly one entry with `IsMain=true`.
- Worktree `Path` matches main repo path.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	r := resp.Repos[0]
	if len(r.Worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %v", r.Worktrees)
	}
	wantMain := absPath(t, filepath.Join(req.Roots[0], "main"))
	if r.Worktrees[0].Path != wantMain {
		t.Fatalf("worktree Path = %q, want %q", r.Worktrees[0].Path, wantMain)
	}
	if !r.Worktrees[0].IsMain {
		t.Fatal("expected IsMain=true for sole main worktree")
	}
}
```