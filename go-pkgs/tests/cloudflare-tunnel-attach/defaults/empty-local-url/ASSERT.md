## Expected

- Attach succeeds.
- `State.Hosts["defaults-url.example.com"].Service` is
  `http://127.0.0.1:6321`.
- Config ingress host rule for that domain uses the same service.

## Side Effects

- Registry and config written under managed dir for TunnelName `team-shared`.

## Errors

- None.

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
		t.Fatalf("Attach error: %v", err)
	}
	if resp == nil || resp.State == nil {
		t.Fatal("nil response or State")
	}
	const host = "defaults-url.example.com"
	const want = "http://127.0.0.1:6321"
	ent := resp.State.Hosts[host]
	if ent == nil {
		t.Fatalf("Hosts missing %q; Hosts=%v", host, resp.State.Hosts)
	}
	if ent.Service != want {
		t.Fatalf("Hosts[%q].Service = %q, want %q", host, ent.Service, want)
	}
	if resp.Config != nil {
		svcs := hostRuleServices(resp.Config)
		if svcs[host] != want {
			t.Fatalf("config service for %q = %q, want %q", host, svcs[host], want)
		}
	}
}
```
