## Expected

- `ResolveGoroot` returns `$InstallDir/go1.19.13`.
- Install hook is not called.
- Stderr does not receive Prompt (dest already exists).

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ResolveGoroot(%q) failed: %v", req.GoVersion, resp.Err)
	}
	want := filepath.Join(req.InstallDir, "go1.19.13")
	if resp.Goroot != want {
		t.Fatalf("ResolveGoroot(%q) = %q, want %q", req.GoVersion, resp.Goroot, want)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); dest already existed", resp.HookVersion, resp.HookDir)
	}
	if resp.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty (Prompt must not run when dest exists)", resp.Stderr)
	}
}
```
