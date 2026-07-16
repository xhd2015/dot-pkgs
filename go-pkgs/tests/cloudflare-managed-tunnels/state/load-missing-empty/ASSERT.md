## Expected

- `LoadTunnelState` returns `err == nil` when `state.json` is missing.
- Returned state is non-nil.
- `Hosts` is non-nil and has length 0.

## Side Effects

- None (read of missing file only).

## Errors

- None (explicit non-error for missing file).

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("LoadTunnelState(missing) error: %v; want nil (empty Hosts policy)", err)
	}
	if resp == nil || resp.State == nil {
		t.Fatal("nil response or State")
	}
	if resp.State.Hosts == nil {
		t.Fatal("Hosts is nil; want non-nil empty map")
	}
	if len(resp.State.Hosts) != 0 {
		t.Fatalf("Hosts len = %d, want 0", len(resp.State.Hosts))
	}
	if resp.DirExists {
		t.Fatal("state.json unexpectedly existed before load")
	}
}
```
