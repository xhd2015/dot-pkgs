## Expected

- `err` is nil.
- `FOO` is present with value `new` (set wins after unset).
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
		t.Fatalf("MergeProcessEnv: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if v, ok := envGet(resp.Env, "FOO"); !ok || v != "new" {
		t.Fatalf("FOO: got (%q, %v), want (new, true)", v, ok)
	}
	if v, ok := envGet(resp.Env, "PATH"); !ok || v != "/bin" {
		t.Fatalf("PATH: got (%q, %v), want (/bin, true)", v, ok)
	}
}
```
