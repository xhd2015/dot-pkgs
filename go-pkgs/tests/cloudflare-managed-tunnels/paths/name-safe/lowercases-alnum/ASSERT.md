## Expected

- `TunnelNameSafe("My-Tunnel")` returns `my-tunnel`.
- Result is stable and filesystem-safe.

## Side Effects

- None (pure function).

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("TunnelNameSafe error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	assertPathEq(t, resp.Path, "my-tunnel")
}
```
