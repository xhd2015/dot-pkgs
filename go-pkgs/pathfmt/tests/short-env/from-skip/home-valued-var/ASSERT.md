## Expected

- `resp.Display` is `~/...` (`TildeHome` form).
- Display is not `$MYHOME...`.

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
	assertNoDollarVar(t, resp.Display, "MYHOME")
	home := mustUserHome(t)
	abs := absPath(t, req.Path)
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q", want, resp.Display)
	}
	if !strings.HasPrefix(resp.Display, "~") {
		t.Fatalf("expected ~ form, got %q", resp.Display)
	}
}
```
