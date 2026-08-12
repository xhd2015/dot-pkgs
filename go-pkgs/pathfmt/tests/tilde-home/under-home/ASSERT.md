## Expected

- `resp.Display` equals `"~" + strings.TrimPrefix(abs, home)` (Short home step).
- `resp.Display` starts with `"~"` and contains the distinctive segment
  `marker-under-home`.
- `resp.Display` does **not** start with the absolute home path (suffix may still
  embed a home-looking segment after `~`).

## Errors

- `err` is nil.

```go
import (
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
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (path=%q home=%q)", want, resp.Display, req.Path, home)
	}
	if !strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("expected ~ prefix, got %q", resp.Display)
	}
	if strings.HasPrefix(resp.Display, home) {
		t.Fatalf("display must not start with absolute home %q: %q", home, resp.Display)
	}
	if !strings.Contains(resp.Display, "marker-under-home") {
		t.Fatalf("expected marker-under-home in display, got %q", resp.Display)
	}
}
```
