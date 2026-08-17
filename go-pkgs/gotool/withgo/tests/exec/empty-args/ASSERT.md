## Expected

- `Exec` runs `env` (empty args).
- Last `GOROOT` in env output is `filepath.Abs(goroot)`.
- Last `PATH` starts with `$abs/bin` plus the path list separator.

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Exec(%q, empty args) failed: %v", req.Goroot, resp.Err)
	}
	abs := absPath(t, req.Goroot)
	gotRoot, ok := lastEnv(resp.Stdout, "GOROOT")
	if !ok {
		t.Fatalf("env output missing GOROOT=\n%s", resp.Stdout)
	}
	if gotRoot != abs {
		t.Fatalf("GOROOT = %q, want %q", gotRoot, abs)
	}
	gotPath, ok := lastEnv(resp.Stdout, "PATH")
	if !ok {
		t.Fatalf("env output missing PATH=\n%s", resp.Stdout)
	}
	wantPrefix := filepath.Join(abs, "bin") + string(os.PathListSeparator)
	if !strings.HasPrefix(gotPath, wantPrefix) {
		t.Fatalf("PATH = %q, want prefix %q", gotPath, wantPrefix)
	}
}
```
