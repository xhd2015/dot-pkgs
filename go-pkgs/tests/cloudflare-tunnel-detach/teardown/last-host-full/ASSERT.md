## Expected

- `err == nil`, `LastStopErr == nil`.
- `DNSDeleteCount >= 1` for `a.example.com`.
- `TunnelDeleteCount >= 1`.
- `ManagedDirRemoved == true`.
- `CredFileRemoved == true`.
- Hosts empty in Response.State (synthetic empty map when dir gone).

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("teardown last-host: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("LastStopErr=%v", resp.LastStopErr)
	}
	const host = "a.example.com"
	if resp.DNSDeleteCount < 1 || resp.DNS == nil || !resp.DNS.hasDeleteFor(host) {
		t.Fatalf("DNS delete missing; count=%d", resp.DNSDeleteCount)
	}
	if resp.TunnelDeleteCount < 1 {
		t.Fatalf("TunnelDeleteCount=%d, want ≥ 1", resp.TunnelDeleteCount)
	}
	if !resp.ManagedDirRemoved {
		t.Fatal("expected ManagedDirRemoved after full teardown")
	}
	if !resp.CredFileRemoved {
		t.Fatal("expected CredFileRemoved after full teardown")
	}
	if resp.State == nil || len(resp.State.Hosts) != 0 {
		t.Fatalf("Hosts should be empty after teardown; state=%+v", resp.State)
	}
}
```
