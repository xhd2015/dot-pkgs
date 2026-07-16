## Expected

- `ManagedTunnelDir(configDir, "")` returns a non-nil error.
- Empty tunnel name must not resolve to a silent default directory.

## Side Effects

- None.

## Errors

- Non-nil error from ManagedTunnelDir (message may mention empty / name).

## Exit Code

- Non-zero API error (`err != nil` from `Run`)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		path := ""
		if resp != nil {
			path = resp.Path
		}
		t.Fatalf("ManagedTunnelDir(empty name) err=nil, path=%q; want error", path)
	}
}
```
