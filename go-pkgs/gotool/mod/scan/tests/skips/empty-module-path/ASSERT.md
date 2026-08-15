## Expected

- `Scan` returns no error.
- Exactly 1 module: the root (`Dir == "."`, `Path == "example.com/root"`).
- No module for `boundary/` (empty module path is not a module).

```go
import (
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
	if resp.Modules[0].Path != "example.com/root" {
		t.Fatalf("root Path = %q, want example.com/root", resp.Modules[0].Path)
	}
}
```
