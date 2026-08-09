## Expected

- `err` is nil.
- `FOO` is present with value `bar`.
- Base keys `PATH` and `HOME` remain with original values.

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
		t.Fatalf("MergeProcessEnv: unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if v, ok := envGet(resp.Env, "FOO"); !ok || v != "bar" {
		t.Fatalf("FOO: got (%q, %v), want (bar, true)", v, ok)
	}
	if v, ok := envGet(resp.Env, "PATH"); !ok || v != "/bin" {
		t.Fatalf("PATH: got (%q, %v), want (/bin, true)", v, ok)
	}
	if v, ok := envGet(resp.Env, "HOME"); !ok || v != "/tmp/home" {
		t.Fatalf("HOME: got (%q, %v), want (/tmp/home, true)", v, ok)
	}
}
```
