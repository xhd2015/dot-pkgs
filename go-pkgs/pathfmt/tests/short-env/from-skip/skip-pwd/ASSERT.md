## Expected

- `resp.Display` is not `$PWD...`.
- Display equals `TildeHome` fallback.

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
	assertNoDollarVar(t, resp.Display, "PWD")
	home := mustUserHome(t)
	abs := absPath(t, req.Path)
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected fallback %q, got %q", want, resp.Display)
	}
}
```
