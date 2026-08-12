## Expected

- `resp.Display` equals the tilde-home form of the absolute path (`~/...`).
- `resp.Display` contains `marker-cwd-and-home`.
- `resp.Display` is **not** a cwd-relative form:
  - not `"."`
  - not equal to `filepath.Rel(cwd, abs)` when that rel has no `".."` prefix
  - does not start with a bare relative segment without `~`

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
	cwdAbs, homeAbs := mustCwdStrictChildOfHome(t)
	abs := absPath(t, req.Path)
	want := expectedTildeHome(abs, homeAbs)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (path=%q cwd=%q)", want, resp.Display, req.Path, cwdAbs)
	}
	if !strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("under cwd+home must use ~ form, got %q", resp.Display)
	}
	if !strings.Contains(resp.Display, "marker-cwd-and-home") {
		t.Fatalf("expected marker-cwd-and-home in display, got %q", resp.Display)
	}
	if strings.HasPrefix(resp.Display, homeAbs) {
		t.Fatalf("display must not start with absolute home %q: %q", homeAbs, resp.Display)
	}
	if resp.Display == "." {
		t.Fatalf("must not be \".\" (cwd form); got %q", resp.Display)
	}
	rel, relErr := filepath.Rel(cwdAbs, abs)
	if relErr == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		if resp.Display == rel {
			t.Fatalf("must not be cwd-relative %q (Short would use this); got tilde? %q", rel, resp.Display)
		}
	}
	if !strings.HasPrefix(resp.Display, "~") && !filepath.IsAbs(resp.Display) {
		t.Fatalf("must not be bare relative form: %q", resp.Display)
	}
}
```
