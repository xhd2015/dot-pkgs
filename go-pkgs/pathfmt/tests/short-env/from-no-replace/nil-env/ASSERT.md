## Expected

- `resp.Display` equals `TildeHome` form (`~/...`).
- Display does not start with `$` (nil env must not load `os.Environ()`).

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
	if req.Env != nil {
		t.Fatalf("test bug: expected nil env, got %#v", req.Env)
	}
	home := mustUserHome(t)
	abs := absPath(t, req.Path)
	want := expectedTildeHome(abs, home)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q", want, resp.Display)
	}
	if strings.HasPrefix(resp.Display, "$") {
		t.Fatalf("nil env must not produce $VAR form: %q", resp.Display)
	}
}
```
