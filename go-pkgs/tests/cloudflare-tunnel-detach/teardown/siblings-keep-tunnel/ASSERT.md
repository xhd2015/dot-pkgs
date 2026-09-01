## Expected

- DNS delete for `a.example.com`.
- `TunnelDeleteCount == 0`.
- `ManagedDirRemoved == false`.
- Hosts only `b.example.com`; connector still up.

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
		t.Fatalf("teardown siblings: %v", err)
	}
	if resp == nil || resp.State == nil {
		t.Fatal("nil response/state")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("LastStopErr=%v", resp.LastStopErr)
	}
	const hostA = "a.example.com"
	const hostB = "b.example.com"
	if _, ok := resp.State.Hosts[hostA]; ok {
		t.Fatalf("Hosts still has %q", hostA)
	}
	if _, ok := resp.State.Hosts[hostB]; !ok {
		t.Fatalf("Hosts missing %q; Hosts=%v", hostB, resp.State.Hosts)
	}
	if resp.DNSDeleteCount < 1 || resp.DNS == nil || !resp.DNS.hasDeleteFor(hostA) {
		t.Fatalf("expected DNS delete for %q; count=%d", hostA, resp.DNSDeleteCount)
	}
	if resp.TunnelDeleteCount != 0 {
		t.Fatalf("TunnelDeleteCount=%d, want 0 when siblings remain", resp.TunnelDeleteCount)
	}
	if resp.ManagedDirRemoved {
		t.Fatal("managed dir must remain when siblings stay")
	}
	if !connectorStillUp(resp) {
		t.Fatal("expected connector still up after partial teardown detach")
	}
}
```
