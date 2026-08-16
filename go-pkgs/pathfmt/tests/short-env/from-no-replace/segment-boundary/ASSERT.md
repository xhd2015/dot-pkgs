## Expected

- `resp.Display` is **not** `$X` or `$X...`.
- Display equals `TildeHome` form of the path (typically absolute `/foo/bar`).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNoDollarVar(t, resp.Display, "X")
	home := mustUserHome(t)
	abs := absPath(t, req.Path)
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected tilde/abs fallback %q, got %q", want, resp.Display)
	}
}
```
