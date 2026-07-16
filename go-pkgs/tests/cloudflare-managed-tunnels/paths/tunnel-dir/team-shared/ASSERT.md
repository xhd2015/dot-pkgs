## Expected

- `ManagedTunnelDir("/tmp/cf-home", "team-shared")` returns
  `/tmp/cf-home/managed-tunnels/team-shared` (filepath join form).
- Path contains the `managed-tunnels` segment and ends with `team-shared`.

## Side Effects

- None (pure function).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("ManagedTunnelDir error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	want := filepath.Join("/tmp/cf-home", "managed-tunnels", "team-shared")
	assertPathEq(t, resp.Path, want)
	if !strings.Contains(resp.Path, "managed-tunnels") {
		t.Fatalf("path %q missing managed-tunnels segment", resp.Path)
	}
}
```
