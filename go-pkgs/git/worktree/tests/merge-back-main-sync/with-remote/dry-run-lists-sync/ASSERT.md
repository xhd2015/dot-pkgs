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
	if !strings.Contains(out, "fetch") {
		t.Fatalf("dry-run output missing fetch:\n%s", out)
	}
	if !strings.Contains(out, "rebase origin/") {
		t.Fatalf("dry-run output missing rebase origin/…:\n%s", out)
	}
	// No land mutation: remote-only and feature files absent on main.
	if _, err := os.Stat(filepath.Join(req.MainRepo, "remote-only.txt")); err == nil {
		t.Fatal("dry-run must not sync/land remote-only.txt onto main")
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "feature.txt")); err == nil {
		t.Fatal("dry-run must not land feature.txt onto main")
	}
}
```
