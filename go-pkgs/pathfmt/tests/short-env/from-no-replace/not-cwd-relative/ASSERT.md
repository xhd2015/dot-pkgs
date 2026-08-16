## Expected

- `resp.Display` equals `TildeHome` form (`~/...`).
- Display is **not** cwd-relative (not `"."`, not a bare relative child).

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
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q", want, resp.Display)
	}
	if resp.Display == "." {
		t.Fatalf("must not be cwd-relative \".\"")
	}
	if !filepath.IsAbs(resp.Display) && !strings.HasPrefix(resp.Display, "~") && !strings.HasPrefix(resp.Display, "$") {
		t.Fatalf("must not be bare relative form: %q", resp.Display)
	}
	if strings.HasPrefix(resp.Display, "doctest-short-env") {
		t.Fatalf("must not be Short-style cwd-relative child: %q", resp.Display)
	}
}
```
