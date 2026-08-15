## Expected

- Both Attach steps succeed (final `err == nil`, two Sessions recorded).
- State Hosts contains `a.example.com` and `b.example.com` with services
  `:7001` and `:7002` respectively.
- Config Ingress length is 3 (2 hosts + 404); both hostnames present; last rule
  only is empty hostname + `http_status:404`.
- RunCount ≥ 2 (first attach start + restart when second host changes ingress).
- RouteDNSCount ≥ 2 (one DNS route per hostname).

## Side Effects

- Single managed dir for TunnelName; one state.json / config.yml rewritten.

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
		t.Fatalf("Attach sequence error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2", len(resp.Sessions))
	}
	if resp.State == nil {
		t.Fatal("nil State")
	}
	if len(resp.State.Hosts) != 2 {
		t.Fatalf("Hosts len = %d, want 2; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	a := resp.State.Hosts["a.example.com"]
	b := resp.State.Hosts["b.example.com"]
	if a == nil || b == nil {
		t.Fatalf("missing host entries: A=%v B=%v", a, b)
	}
	if a.Service != "http://127.0.0.1:7001" {
		t.Fatalf("A.Service = %q, want http://127.0.0.1:7001", a.Service)
	}
	if b.Service != "http://127.0.0.1:7002" {
		t.Fatalf("B.Service = %q, want http://127.0.0.1:7002", b.Service)
	}

	if resp.Config == nil {
		t.Fatal("nil Config")
	}
	if len(resp.Config.Ingress) != 3 {
		t.Fatalf("Ingress len = %d, want 3 (2 hosts + 404)", len(resp.Config.Ingress))
	}
	svcs := hostRuleServices(resp.Config)
	if svcs["a.example.com"] != "http://127.0.0.1:7001" {
		t.Fatalf("config A service = %q", svcs["a.example.com"])
	}
	if svcs["b.example.com"] != "http://127.0.0.1:7002" {
		t.Fatalf("config B service = %q", svcs["b.example.com"])
	}
	if len(svcs) != 2 {
		t.Fatalf("host rule count = %d, want 2; %v", len(svcs), svcs)
	}
	last, ok := lastIngress(resp.Config)
	if !ok {
		t.Fatal("missing last ingress")
	}
	if last.Hostname != "" || last.Service != "http_status:404" {
		t.Fatalf("last ingress = %+v, want empty host + http_status:404", last)
	}
	// Ensure 404 is not duplicated as a non-last empty-host rule.
	for i, r := range resp.Config.Ingress {
		if r.Hostname == "" && i != len(resp.Config.Ingress)-1 {
			t.Fatalf("empty-hostname rule at index %d; catch-all must be last only", i)
		}
	}

	if resp.RunCount < 2 {
		t.Fatalf("RunCount = %d, want ≥ 2 (start + restart on second host)", resp.RunCount)
	}
	if resp.RouteDNSCount < 2 {
		t.Fatalf("RouteDNSCount = %d, want ≥ 2", resp.RouteDNSCount)
	}
}
```
