## Expected

- `Scan` returns no error.
- Exactly 1 module: the root (`Dir == "."`).
- No module whose `Dir` starts with `testdata` (the `testdata/` subtree is pruned).

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
		if strings.HasPrefix(m.Dir, "testdata") {
			t.Fatalf("testdata subtree was not skipped: %+v", m)
		}
	}
}
```
