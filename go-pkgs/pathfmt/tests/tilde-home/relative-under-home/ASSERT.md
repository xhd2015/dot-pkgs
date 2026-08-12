## Expected

- After Abs under process cwd (under home), `resp.Display` equals the tilde-home
  form (`"~" + TrimPrefix(abs, home)`).
- `resp.Display` contains `marker-relative-under-home`.
- `resp.Display` does **not** start with the absolute home path (suffix may still
  embed a home-looking segment after `~`).
- `resp.Display` is **not** the bare relative input unchanged when Abs places it
  under home (must prefer `~/...`).

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
	if filepath.IsAbs(req.Path) {
		t.Fatalf("leaf requires relative input, got %q", req.Path)
	}
	_, homeAbs := mustCwdUnderHome(t)
	abs := absPath(t, req.Path)
	if !isUnderHome(abs, homeAbs) {
		t.Fatalf("expected Abs(%q)=%q under home %q", req.Path, abs, homeAbs)
	}
	want := expectedTildeHome(abs, homeAbs)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (rel path=%q abs=%q)", want, resp.Display, req.Path, abs)
	}
	if !strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("relative-under-home must use ~ form, got %q", resp.Display)
	}
	if !strings.Contains(resp.Display, "marker-relative-under-home") {
		t.Fatalf("expected marker-relative-under-home in display, got %q", resp.Display)
	}
	if strings.HasPrefix(resp.Display, homeAbs) {
		t.Fatalf("display must not start with absolute home %q: %q", homeAbs, resp.Display)
	}
	if resp.Display == req.Path {
		t.Fatalf("must not leave relative input unchanged when under home: %q", resp.Display)
	}
}
```
