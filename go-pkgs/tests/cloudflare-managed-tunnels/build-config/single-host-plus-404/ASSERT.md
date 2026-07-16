## Expected

- Build succeeds.
- Config.Tunnel equals StateIn.TunnelID; CredentialsFile matches state.
- Ingress length is 2.
- First rule: Hostname `app.example.com`, Service `http://127.0.0.1:6321`.
- Last rule: empty Hostname, Service `http_status:404`.

## Side Effects

- None (pure build).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("BuildConfigFromState error: %v", err)
	}
	if resp == nil || resp.Config == nil {
		t.Fatal("nil response or Config")
	}
	cfg := resp.Config
	if cfg.Tunnel != req.StateIn.TunnelID {
		t.Fatalf("Tunnel = %q, want %q", cfg.Tunnel, req.StateIn.TunnelID)
	}
	if cfg.CredentialsFile != req.StateIn.CredentialsFile {
		t.Fatalf("CredentialsFile = %q, want %q", cfg.CredentialsFile, req.StateIn.CredentialsFile)
	}
	if len(cfg.Ingress) != 2 {
		t.Fatalf("Ingress len = %d, want 2 (host + 404)", len(cfg.Ingress))
	}
	r0 := cfg.Ingress[0]
	if r0.Hostname != "app.example.com" {
		t.Fatalf("Ingress[0].Hostname = %q, want app.example.com", r0.Hostname)
	}
	if r0.Service != "http://127.0.0.1:6321" {
		t.Fatalf("Ingress[0].Service = %q, want http://127.0.0.1:6321", r0.Service)
	}
	last, ok := lastIngress(cfg)
	if !ok {
		t.Fatal("missing last ingress")
	}
	if last.Hostname != "" {
		t.Fatalf("last Hostname = %q, want empty catch-all", last.Hostname)
	}
	if last.Service != "http_status:404" {
		t.Fatalf("last Service = %q, want http_status:404", last.Service)
	}
}
```
