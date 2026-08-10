## Expected

- `err` is nil.
- Environ map equals base (unset of missing key is no-op).
- `MISSING` remains absent.

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
		t.Fatalf("MergeProcessEnv: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !envMapsEqual(resp.Env, req.Base) {
		t.Fatalf("env map mismatch:\n  got  %#v\n  want %#v", envAsMap(resp.Env), envAsMap(req.Base))
	}
	if envHas(resp.Env, "MISSING") {
		t.Fatalf("MISSING should stay absent, env=%v", resp.Env)
	}
}
```
