## Expected

- Attach succeeds (`err == nil`).
- Session Domain is `a.example.com`; ConfigPath is under ManagedDir.
- State has exactly one Hosts entry for `a.example.com` with service
  `http://127.0.0.1:6321`.
- State TunnelName / TunnelID / CredentialsFile are non-empty.
- `config.yml` exists at ManagedDir/config.yml.
- Parsed config Ingress length is 2: host rule then catch-all
  `http_status:404` with empty hostname.
- Session PublicBaseURL / `https://a.example.com` when available.

## Side Effects

- Creates managed tunnel directory contents under ConfigDir.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach error: %v", err)
	}
	if resp == nil || resp.Session == nil {
		t.Fatal("nil response or Session")
	}
	const host = "a.example.com"
	const svc = "http://127.0.0.1:6321"

	if resp.Session.Domain != host {
		t.Fatalf("Session.Domain = %q, want %q", resp.Session.Domain, host)
	}
	if resp.ManagedDir == "" {
		t.Fatal("ManagedDir empty")
	}
	if resp.ConfigPath == "" {
		t.Fatal("ConfigPath empty")
	}
	if _, serr := os.Stat(resp.ConfigPath); serr != nil {
		t.Fatalf("config.yml missing at %s: %v", resp.ConfigPath, serr)
	}
	// ConfigPath should live under managed dir.
	if filepath.Dir(resp.ConfigPath) != resp.ManagedDir {
		t.Fatalf("ConfigPath dir %q != ManagedDir %q", filepath.Dir(resp.ConfigPath), resp.ManagedDir)
	}
	if resp.Session.ConfigPath != "" && resp.Session.ConfigPath != resp.ConfigPath {
		// Allow Session.ConfigPath to match managed config path when set.
		if filepath.Clean(resp.Session.ConfigPath) != filepath.Clean(resp.ConfigPath) {
			t.Fatalf("Session.ConfigPath = %q, want %q", resp.Session.ConfigPath, resp.ConfigPath)
		}
	}

	if resp.State == nil {
		t.Fatal("nil State")
	}
	if len(resp.State.Hosts) != 1 {
		t.Fatalf("Hosts len = %d, want 1; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	ent := resp.State.Hosts[host]
	if ent == nil {
		t.Fatalf("Hosts missing %q", host)
	}
	if ent.Service != svc {
		t.Fatalf("Hosts[%q].Service = %q, want %q", host, ent.Service, svc)
	}
	if resp.State.TunnelName == "" {
		t.Fatal("State.TunnelName empty")
	}
	if resp.State.TunnelID == "" {
		t.Fatal("State.TunnelID empty")
	}
	if resp.State.CredentialsFile == "" {
		t.Fatal("State.CredentialsFile empty")
	}

	if resp.Config == nil {
		t.Fatal("nil Config (config.yml missing or unreadable)")
	}
	if len(resp.Config.Ingress) != 2 {
		t.Fatalf("Ingress len = %d, want 2 (host + 404)", len(resp.Config.Ingress))
	}
	svcs := hostRuleServices(resp.Config)
	if svcs[host] != svc {
		t.Fatalf("config host service = %q, want %q", svcs[host], svc)
	}
	last, ok := lastIngress(resp.Config)
	if !ok {
		t.Fatal("missing last ingress")
	}
	if last.Hostname != "" {
		t.Fatalf("last Hostname = %q, want empty", last.Hostname)
	}
	if last.Service != "http_status:404" {
		t.Fatalf("last Service = %q, want http_status:404", last.Service)
	}

	if pub := resp.Session.PublicBaseURL(); pub != "" && pub != "https://"+host {
		t.Fatalf("PublicBaseURL = %q, want https://%s", pub, host)
	}
}
```
