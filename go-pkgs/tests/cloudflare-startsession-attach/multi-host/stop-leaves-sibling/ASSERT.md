## Expected

- StartSession A+B + Stop(A) complete with final `err == nil` and
  `LastStopErr == nil`.
- State Hosts has exactly one entry: `b.example.com` with service
  `http://127.0.0.1:7002`.
- Hosts must **not** contain `a.example.com`.
- When Config is readable: only B host rule + 404 last.

## Side Effects

- Single managed dir rewritten; sibling B retained after Stop(A).

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
		t.Fatalf("StartSession/Stop sequence error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("Stop(A) error: %v", resp.LastStopErr)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2 (StartSession A and B)", len(resp.Sessions))
	}
	if resp.State == nil {
		t.Fatal("nil State after Stop(A) (managed registry missing — StartSession must Attach)")
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

	if resp.Config != nil {
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
		last, ok := lastIngress(resp.Config)
		if !ok {
			t.Fatal("missing last ingress")
		}
		if last.Hostname != "" || last.Service != "http_status:404" {
			t.Fatalf("last ingress = %+v, want empty host + http_status:404", last)
		}
	}
}
```
