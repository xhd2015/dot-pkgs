## Expected

- Attach succeeds.
- `Session.TunnelName` equals `cloudflare.DefaultTunnelName`.
- Loaded state `TunnelName` equals `DefaultTunnelName`.
- Managed dir path ends with `managed-tunnels/` + `TunnelNameSafe(DefaultTunnelName)`.

## Side Effects

- Writes `state.json` / `config.yml` under the default tunnel managed dir.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach error: %v", err)
	}
	if resp == nil || resp.Session == nil {
		t.Fatal("nil response or Session")
	}
	want := cloudflare.DefaultTunnelName
	if resp.Session.TunnelName != want {
		t.Fatalf("Session.TunnelName = %q, want %q", resp.Session.TunnelName, want)
	}
	if resp.State == nil {
		t.Fatal("nil State after Attach")
	}
	if resp.State.TunnelName != want {
		t.Fatalf("State.TunnelName = %q, want %q", resp.State.TunnelName, want)
	}
	safe := cloudflare.TunnelNameSafe(want)
	suffix := filepath.Join("managed-tunnels", safe)
	if !strings.HasSuffix(filepath.Clean(resp.ManagedDir), suffix) {
		t.Fatalf("ManagedDir %q should end with %q", resp.ManagedDir, suffix)
	}
}
```
