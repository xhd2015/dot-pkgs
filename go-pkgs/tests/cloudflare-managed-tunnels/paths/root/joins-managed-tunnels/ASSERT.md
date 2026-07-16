## Expected

- `ManagedTunnelsRoot("/tmp/cf")` returns `/tmp/cf/managed-tunnels`
  (filepath join of config dir and `managed-tunnels`).
- No trailing extra segments.

## Side Effects

- None (pure function).

## Errors

- None.

## Exit Code

- 0 (`err == nil` from `Run`)

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("ManagedTunnelsRoot error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	want := filepath.Join("/tmp/cf", "managed-tunnels")
	assertPathEq(t, resp.Path, want)
}
```
