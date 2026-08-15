## Expected

- `Scan` returns no error.
- Exactly 1 module: the root (`Dir == "."`).
- No module whose `Dir` starts with `sub/build` — the gitignored subtree
  (matched by nested `sub/.gitignore`) is pruned.

## Exit Code

- (N/A — library test, no CLI)

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
	for _, m := range resp.Modules {
		if strings.HasPrefix(m.Dir, "sub/build") {
			t.Fatalf("nested-gitignored dir was not skipped: %+v", m)
		}
	}
}
```
