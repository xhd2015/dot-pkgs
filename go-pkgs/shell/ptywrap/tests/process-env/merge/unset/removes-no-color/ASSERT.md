## Expected

- `err` is nil.
- `NO_COLOR` is absent.
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
	if envHas(resp.Env, "NO_COLOR") {
		t.Fatalf("NO_COLOR should be absent, env=%v", resp.Env)
	}
	if v, ok := envGet(resp.Env, "PATH"); !ok || v != "/bin" {
		t.Fatalf("PATH: got (%q, %v), want (/bin, true)", v, ok)
	}
}
```
