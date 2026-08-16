## Expected

- `resp.Display` equals the absolute path.
- Display does not start with `~` or `$`.

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
		t.Skipf("temp path %q unexpectedly under home %q", abs, home)
	}
	if resp.Display != abs {
		t.Fatalf("expected absolute %q, got %q", abs, resp.Display)
	}
	if strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("outside-home must not use ~: %q", resp.Display)
	}
	if strings.HasPrefix(resp.Display, "$") {
		t.Fatalf("outside-home with empty env must not use $VAR: %q", resp.Display)
	}
	if !filepath.IsAbs(resp.Display) {
		t.Fatalf("must be absolute, got %q", resp.Display)
	}
}
```
