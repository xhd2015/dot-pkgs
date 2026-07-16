## Expected

- Final `err == nil` and `LastStopErr == nil` (no-op success preferred).
- State Hosts still exactly `{b.example.com}` with service
  `http://127.0.0.1:7002`.
- Hosts does not gain or lose unexpected keys.
- Config still has host rule for B + 404 last.
- Connector still logical-up (`ConnectorPID > 0` or run count from Attach ≥ 1
  with PID not forced to 0 by missing-host Detach).

## Side Effects

- No destructive rewrite of sibling host B.

## Errors

- None (prefer success).

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Detach(missing domain) error: %v; want no-op success", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("LastStopErr = %v, want nil (no-op success)", resp.LastStopErr)
	}
	if resp.State == nil {
		t.Fatal("nil State")
	}

	const hostA = "a.example.com"
	const hostB = "b.example.com"
	const svcB = "http://127.0.0.1:7002"

	if _, ok := resp.State.Hosts[hostA]; ok {
		t.Fatalf("Hosts unexpectedly contains missing domain %q", hostA)
	}
	if len(resp.State.Hosts) != 1 {
		t.Fatalf("Hosts len = %d, want 1 (B unchanged); Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	b := resp.State.Hosts[hostB]
	if b == nil {
		t.Fatalf("Hosts missing sibling %q after Detach(missing); Hosts=%v", hostB, resp.State.Hosts)
	}
	if b.Service != svcB {
		t.Fatalf("Hosts[%q].Service = %q, want %q", hostB, b.Service, svcB)
	}

	// Missing detach must not tear down the only real host's connector.
	if resp.State.ConnectorPID <= 0 {
		t.Fatalf("ConnectorPID = %d after Detach(missing), want > 0 (sibling B still attached); RunCount=%d",
			resp.State.ConnectorPID, resp.RunCount)
	}

	if resp.Config != nil {
		svcs := hostRuleServices(resp.Config)
		if svcs[hostB] != svcB {
			t.Fatalf("config B service = %q, want %q", svcs[hostB], svcB)
		}
		if _, ok := svcs[hostA]; ok {
			t.Fatalf("config has unexpected host rule for %q", hostA)
		}
	}
}
```
