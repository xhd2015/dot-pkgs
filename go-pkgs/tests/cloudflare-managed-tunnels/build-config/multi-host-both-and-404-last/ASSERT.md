## Expected

- Build succeeds.
- Ingress length is 3 (2 host rules + 404).
- Hostnames `a.example.com` and `b.example.com` both present with correct services.
- Exactly one catch-all: last entry has empty Hostname and `http_status:404`.
- No non-last rule uses `http_status:404` with empty hostname (404 last only).

## Side Effects

- None (pure build).

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
		t.Fatalf("BuildConfigFromState error: %v", err)
	}
	if resp == nil || resp.Config == nil {
		t.Fatal("nil response or Config")
	}
	cfg := resp.Config
	if len(cfg.Ingress) != 3 {
		t.Fatalf("Ingress len = %d, want 3 (2 hosts + 404)", len(cfg.Ingress))
	}

	byHost := map[string]string{}
	for i, r := range cfg.Ingress {
		if r.Hostname == "" {
			if i != len(cfg.Ingress)-1 {
				t.Fatalf("empty-hostname rule at index %d; catch-all must be last only", i)
			}
			continue
		}
		byHost[r.Hostname] = r.Service
	}
	if byHost["a.example.com"] != "http://127.0.0.1:7001" {
		t.Fatalf("a.example.com service = %q, want http://127.0.0.1:7001", byHost["a.example.com"])
	}
	if byHost["b.example.com"] != "http://127.0.0.1:7002" {
		t.Fatalf("b.example.com service = %q, want http://127.0.0.1:7002", byHost["b.example.com"])
	}
	if len(byHost) != 2 {
		t.Fatalf("host rule count = %d, want 2; byHost=%v", len(byHost), byHost)
	}

	last, ok := lastIngress(cfg)
	if !ok {
		t.Fatal("missing last ingress")
	}
	if last.Hostname != "" {
		t.Fatalf("last Hostname = %q, want empty", last.Hostname)
	}
	if last.Service != "http_status:404" {
		t.Fatalf("last Service = %q, want http_status:404", last.Service)
	}
}
```
