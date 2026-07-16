## Expected

- Attach + Stop(A) complete with final `err == nil` (and `LastStopErr == nil`).
- State Hosts has exactly one entry: `b.example.com` with service
  `http://127.0.0.1:7002`.
- Hosts must **not** contain `a.example.com`.
- Config Ingress: one host rule for B + catch-all `http_status:404` last
  (Ingress length 2); no hostname rule for A.
- Connector still logical-up: `State.ConnectorPID > 0` **or**
  `RunCount > RunCountAfterAttach` (restart with remaining hosts).

## Side Effects

- Single managed dir rewritten in place; sibling B retained.

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach/Stop sequence error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("Stop(A) error: %v", resp.LastStopErr)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2 (Attach A and B)", len(resp.Sessions))
	}
	if resp.State == nil {
		t.Fatal("nil State after Stop(A)")
	}

	const hostA = "a.example.com"
	const hostB = "b.example.com"
	const svcB = "http://127.0.0.1:7002"

	if _, ok := resp.State.Hosts[hostA]; ok {
		t.Fatalf("Hosts still contains %q after Stop(A); Hosts=%v", hostA, resp.State.Hosts)
	}
	if len(resp.State.Hosts) != 1 {
		t.Fatalf("Hosts len = %d, want 1 (only B); Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	b := resp.State.Hosts[hostB]
	if b == nil {
		t.Fatalf("Hosts missing %q after Stop(A); Hosts=%v", hostB, resp.State.Hosts)
	}
	if b.Service != svcB {
		t.Fatalf("Hosts[%q].Service = %q, want %q", hostB, b.Service, svcB)
	}

	if resp.Config == nil {
		t.Fatal("nil Config (config.yml missing or unreadable after Stop)")
	}
	svcs := hostRuleServices(resp.Config)
	if _, ok := svcs[hostA]; ok {
		t.Fatalf("config still has host rule for %q; svcs=%v", hostA, svcs)
	}
	if svcs[hostB] != svcB {
		t.Fatalf("config B service = %q, want %q; svcs=%v", svcs[hostB], svcB, svcs)
	}
	if len(svcs) != 1 {
		t.Fatalf("host rule count = %d, want 1; %v", len(svcs), svcs)
	}
	if len(resp.Config.Ingress) != 2 {
		t.Fatalf("Ingress len = %d, want 2 (B + 404)", len(resp.Config.Ingress))
	}
	last, ok := lastIngress(resp.Config)
	if !ok {
		t.Fatal("missing last ingress")
	}
	if last.Hostname != "" || last.Service != "http_status:404" {
		t.Fatalf("last ingress = %+v, want empty host + http_status:404", last)
	}
	for i, r := range resp.Config.Ingress {
		if r.Hostname == "" && i != len(resp.Config.Ingress)-1 {
			t.Fatalf("empty-hostname rule at index %d; catch-all must be last only", i)
		}
	}

	if !connectorStillUp(resp) {
		t.Fatalf("connector not up after partial detach: ConnectorPID=%d RunCount=%d RunCountAfterAttach=%d",
			resp.State.ConnectorPID, resp.RunCount, resp.RunCountAfterAttach)
	}
}
```
