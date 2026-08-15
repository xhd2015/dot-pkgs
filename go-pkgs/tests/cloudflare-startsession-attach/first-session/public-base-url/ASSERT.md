## Expected

- StartSession succeeds (`err == nil`).
- Session Domain is `a.example.com`.
- `Session.PublicBaseURL()` is exactly `https://a.example.com`.
- ManagedDir is non-empty; `ConfigPath` is `ManagedDir/config.yml` and exists
  (or Session.ConfigPath is under ManagedDir).
- State has exactly one Hosts entry for `a.example.com` with service
  `http://127.0.0.1:6321`.
- State TunnelName / TunnelID are non-empty when state is written via Attach.

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
		t.Fatalf("StartSession error: %v", err)
	}
	if resp == nil || resp.Session == nil {
		t.Fatal("nil response or Session")
	}

	const host = "a.example.com"
	const svc = "http://127.0.0.1:6321"
	const wantURL = "https://" + host

	if resp.Session.Domain != host {
		t.Fatalf("Session.Domain = %q, want %q", resp.Session.Domain, host)
	}
	if pub := resp.Session.PublicBaseURL(); pub != wantURL {
		t.Fatalf("PublicBaseURL = %q, want %q", pub, wantURL)
	}

	if resp.ManagedDir == "" {
		t.Fatal("ManagedDir empty")
	}
	// Config must live under managed dir (StartSession → Attach path).
	configOK := false
	if resp.ConfigPath != "" {
		if filepath.Dir(resp.ConfigPath) == resp.ManagedDir {
			if _, serr := os.Stat(resp.ConfigPath); serr == nil {
				configOK = true
			}
		}
	}
	if resp.Session.ConfigPath != "" {
		if filepath.Dir(filepath.Clean(resp.Session.ConfigPath)) == filepath.Clean(resp.ManagedDir) {
			if _, serr := os.Stat(resp.Session.ConfigPath); serr == nil {
				configOK = true
			}
		}
	}
	if !configOK {
		t.Fatalf("managed config.yml missing under ManagedDir=%q ConfigPath=%q Session.ConfigPath=%q (StartSession must use Attach managed path)",
			resp.ManagedDir, resp.ConfigPath, resp.Session.ConfigPath)
	}

	if resp.State == nil {
		t.Fatal("nil State (managed state.json missing — StartSession must Attach)")
	}
	if len(resp.State.Hosts) != 1 {
		t.Fatalf("Hosts len = %d, want 1; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	ent := resp.State.Hosts[host]
	if ent == nil {
		t.Fatalf("Hosts missing %q; Hosts=%v", host, resp.State.Hosts)
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
}
```
