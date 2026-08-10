## Expected

- `err` is nil.
- `TERM` remains `screen-256color` (not replaced with `xterm-256color`).
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
	if err != nil {
		t.Fatalf("spawn env: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if v, ok := envGet(resp.Env, "TERM"); !ok || v != "screen-256color" {
		t.Fatalf("TERM: got (%q, %v), want (screen-256color, true) — must not force xterm-256color", v, ok)
	}
	if v, ok := envGet(resp.Env, "PATH"); !ok || v != "/bin" {
		t.Fatalf("PATH: got (%q, %v), want (/bin, true)", v, ok)
	}
}
```
