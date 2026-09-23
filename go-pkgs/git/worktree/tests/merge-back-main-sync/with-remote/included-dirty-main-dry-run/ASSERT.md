# Assert

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Err != "" {
		t.Fatalf("unexpected error: %s", resp.Err)
	}
	if resp.Action != "dry-run" {
		t.Fatalf("Action=%q want dry-run", resp.Action)
	}
	out := ""
	if req.Stdout != nil {
		out = req.Stdout.String()
	}
	if strings.Contains(out, "fetch") || strings.Contains(out, "rebase") {
		t.Fatalf("included dry-run must not list main-sync:\n%s", out)
	}
	if !strings.Contains(out, "worktree remove") {
		t.Fatalf("included dry-run missing worktree remove:\n%s", out)
	}
	if !strings.Contains(out, "branch -D") {
		t.Fatalf("included dry-run missing branch -D:\n%s", out)
	}
	if _, err := os.Stat(req.SourcePath); err != nil {
		t.Fatalf("dry-run must keep worktree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "remote-only.txt")); err == nil {
		t.Fatal("dry-run must not sync remote-only.txt onto main")
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "dirty-main.txt")); err != nil {
		t.Fatalf("dirty-main.txt should remain: %v", err)
	}
}
```
