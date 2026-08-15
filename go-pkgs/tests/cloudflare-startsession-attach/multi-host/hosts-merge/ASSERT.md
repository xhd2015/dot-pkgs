## Expected

- Both StartSession steps succeed (final `err == nil`, two Sessions recorded).
- State Hosts contains `a.example.com` and `b.example.com` with services
  `:7001` and `:7002` respectively.
- Hosts length is exactly 2.
- Prefer config both hostnames + 404 last when Config is readable (soft if
  Config nil — Hosts is the hard requirement).

## Side Effects

- Single managed dir for TunnelName; one state.json rewritten with both hosts.

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
		t.Fatalf("StartSession sequence error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2", len(resp.Sessions))
	}
	if resp.State == nil {
		t.Fatal("nil State (managed state.json missing — StartSession must Attach)")
	}
	if len(resp.State.Hosts) != 2 {
		t.Fatalf("Hosts len = %d, want 2; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	a := resp.State.Hosts["a.example.com"]
	b := resp.State.Hosts["b.example.com"]
	if a == nil || b == nil {
		t.Fatalf("missing host entries: A=%v B=%v Hosts=%v", a, b, resp.State.Hosts)
	}
	if a.Service != "http://127.0.0.1:7001" {
		t.Fatalf("A.Service = %q, want http://127.0.0.1:7001", a.Service)
	}
	if b.Service != "http://127.0.0.1:7002" {
		t.Fatalf("B.Service = %q, want http://127.0.0.1:7002", b.Service)
	}

	// Optional but preferred after Attach rewrite: multi-host config shape.
	if resp.Config != nil {
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
	} else {
		t.Logf("note: Config nil (config.yml not under managed dir yet); Hosts asserted")
	}
}
```
