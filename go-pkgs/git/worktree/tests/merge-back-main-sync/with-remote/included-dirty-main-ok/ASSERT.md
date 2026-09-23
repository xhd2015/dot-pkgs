# Assert

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}
	if resp.Err != "" {
		t.Fatalf("unexpected error: %s", resp.Err)
	}
	if resp.Action != "removed" {
		t.Fatalf("Action=%q want removed", resp.Action)
	}
	if _, err := os.Stat(req.SourcePath); !os.IsNotExist(err) {
		t.Fatalf("worktree should be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "feature.txt")); err != nil {
		t.Fatalf("feature.txt should remain on main: %v", err)
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "remote-only.txt")); err == nil {
		t.Fatal("remote-only.txt must not land; main-sync should have been skipped")
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "dirty-main.txt")); err != nil {
		t.Fatalf("dirty-main.txt should remain uncommitted on main: %v", err)
	}
}
```
