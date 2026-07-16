## Expected

- Sequence succeeds.
- State Hosts has exactly one entry `a.example.com` with Service
  `http://127.0.0.1:6322` (updated, not duplicated).
- Config host rule for A uses `:6322`; still ends with `http_status:404`.
- RunCount ≥ 2 (restart on service change).

## Side Effects

- state.json / config.yml rewritten in place under the same managed dir.

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach sequence error: %v", err)
	}
	if resp == nil || resp.State == nil {
		t.Fatal("nil response or State")
	}
	const host = "a.example.com"
	const want = "http://127.0.0.1:6322"
	if len(resp.State.Hosts) != 1 {
		t.Fatalf("Hosts len = %d, want 1; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	ent := resp.State.Hosts[host]
	if ent == nil {
		t.Fatalf("Hosts missing %q", host)
	}
	if ent.Service != want {
		t.Fatalf("Hosts[%q].Service = %q, want %q", host, ent.Service, want)
	}
	if resp.Config != nil {
		svcs := hostRuleServices(resp.Config)
		if svcs[host] != want {
			t.Fatalf("config service = %q, want %q", svcs[host], want)
		}
		last, ok := lastIngress(resp.Config)
		if !ok || last.Service != "http_status:404" || last.Hostname != "" {
			t.Fatalf("last ingress = %+v, want 404 catch-all", last)
		}
	}
	if resp.RunCount < 2 {
		t.Fatalf("RunCount = %d, want ≥ 2 (restart on LocalURL change)", resp.RunCount)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2", len(resp.Sessions))
	}
}
```
