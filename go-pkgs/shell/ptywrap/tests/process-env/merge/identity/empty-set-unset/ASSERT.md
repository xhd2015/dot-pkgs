## Expected

- `err` is nil.
- `resp.Env` has the same KEY→value map as `req.Base`.
- `TERM` is still absent (pure merge does not force spawn TERM default).
- `FOO` remains `keep`.

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
	if err != nil {
		t.Fatalf("MergeProcessEnv: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !envMapsEqual(resp.Env, req.Base) {
		t.Fatalf("env map mismatch:\n  got  %#v\n  want %#v", envAsMap(resp.Env), envAsMap(req.Base))
	}
	if envHas(resp.Env, "TERM") {
		t.Fatalf("pure merge must not inject TERM; got %q", resp.Env)
	}
	if v, ok := envGet(resp.Env, "FOO"); !ok || v != "keep" {
		t.Fatalf("FOO: got (%q, %v), want (keep, true)", v, ok)
	}
}
```
