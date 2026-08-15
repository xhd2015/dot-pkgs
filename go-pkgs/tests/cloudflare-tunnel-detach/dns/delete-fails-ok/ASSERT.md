## Expected

- Final `err == nil` (DNS failure must not fail Detach).
- `LastStopErr == nil`.
- State Hosts does not contain `a.example.com` (host still detached).
- Prefer Hosts empty after sole-host detach.
- `DNSDeleteCount >= 1` and fake deleter recorded hostname `a.example.com`.
- `ConnectorPID == 0` after last host removed (same as last-host sole stop).

## Side Effects

- DNS delete attempted once (or more); failure only logged in production.

## Errors

- None at API boundary (DNS error swallowed / best-effort).

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
		t.Fatalf("Detach with failing DNS returned error: %v (want nil / best-effort)", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("LastStopErr = %v, want nil when only DNS fails", resp.LastStopErr)
	}
	if resp.State == nil {
		t.Fatal("nil State after Detach")
	}
	const host = "a.example.com"
	if _, ok := resp.State.Hosts[host]; ok {
		t.Fatalf("Hosts still has %q after Detach despite DNS failure; Hosts=%v", host, resp.State.Hosts)
	}
	if len(resp.State.Hosts) != 0 {
		t.Fatalf("Hosts len = %d, want 0; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	if resp.DNSDeleteCount < 1 {
		t.Fatalf("DNSDeleteCount = %d, want ≥ 1 (DeleteHostname attempted)", resp.DNSDeleteCount)
	}
	if resp.DNS == nil || !resp.DNS.hasDeleteFor(host) {
		t.Fatalf("DNS deleter missing DeleteHostname(%q); count=%d", host, resp.DNSDeleteCount)
	}
	if resp.State.ConnectorPID != 0 {
		t.Fatalf("ConnectorPID = %d, want 0 after last host detach", resp.State.ConnectorPID)
	}
}
```
