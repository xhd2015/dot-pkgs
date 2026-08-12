## Expected

- `resp.Display` equals the absolute form of `req.Path`.
- `resp.Display` does **not** start with `"~"`.
- `resp.Display` is absolute (not cwd-relative).

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	home := mustUserHome(t)
	abs := absPath(t, req.Path)
	if isUnderHome(abs, home) {
		t.Skipf("temp path %q unexpectedly under home %q; cannot assert outside-home", abs, home)
	}
	if resp.Display != abs {
		t.Fatalf("expected absolute %q, got %q", abs, resp.Display)
	}
	if strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("outside-home path must not use ~ prefix: %q", resp.Display)
	}
	if !filepath.IsAbs(resp.Display) {
		t.Fatalf("outside-home path must be absolute, got %q", resp.Display)
	}
	if resp.Display == "." || strings.HasPrefix(resp.Display, "."+string(filepath.Separator)) {
		t.Fatalf("outside-home path must not be cwd-relative: %q", resp.Display)
	}
}
```
