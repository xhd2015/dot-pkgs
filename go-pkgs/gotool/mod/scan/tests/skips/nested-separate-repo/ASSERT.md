## Expected

- `Scan` returns no error.
- Exactly 1 module: the root (`Dir == "."`).
- No module whose `Dir` is `ext` or starts with `ext/` — the nested separate repo's whole
  subtree (including its own `go.mod`) is skipped.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Scan(%q) failed: %v", req.RootDir, resp.Err)
	}
	if len(resp.Modules) != 1 {
		t.Fatalf("Scan returned %d modules, want 1: %+v", len(resp.Modules), resp.Modules)
	}
	if resp.Modules[0].Dir != "." {
		t.Fatalf("only module Dir = %q, want \".\"", resp.Modules[0].Dir)
	}
	for _, m := range resp.Modules {
		if m.Dir == "ext" || strings.HasPrefix(m.Dir, "ext/") {
			t.Fatalf("nested separate repo was not skipped: %+v", m)
		}
	}
}
```
