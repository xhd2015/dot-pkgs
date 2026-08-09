## Expected

- `err` is nil.
- `TERM` is `xterm-256color` (empty replaced).
- `PATH` remains `/bin`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("spawn env: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if v, ok := envGet(resp.Env, "TERM"); !ok || v != "xterm-256color" {
		t.Fatalf("TERM: got (%q, %v), want (xterm-256color, true)", v, ok)
	}
	if v, ok := envGet(resp.Env, "PATH"); !ok || v != "/bin" {
		t.Fatalf("PATH: got (%q, %v), want (/bin, true)", v, ok)
	}
}
```
