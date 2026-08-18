# Scenario

**Feature**: Detach / Session.Stop one hostname from managed named tunnel

```
# production
Detach(domain, tunnelName, configDir, runner, dnsDeleter)
  -> flock managed dir
  -> remove Hosts[domain]
  -> BuildConfigFromState -> config.yml
  -> if hosts remain: runner Exec tunnel run (restart)
  -> if hosts empty: ConnectorPID=0; connector down
  -> best-effort DeleteDNS(domain)

# Session.Stop (managed Attach session)
  -> prefer Detach with session Domain / tunnel / runner / dnsDeleter

# doctest path (this tree — package API, network free)
Mode=partial-detach|last-host|dns|missing-host
  -> Attach sequence with fake CommandRunner
  -> Session.Stop and/or Detach
  -> Response.State / Config / RunCount / ConnectorPID / DNSDeleteCount
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`
  (local package in this go-pkgs module; no replace).
- **Coverage backfill / GREEN expected:** `Detach` / managed `Session.Stop` present.
- `Attach` available for setup.
- Managed helpers for post-condition reads: `ManagedTunnelDir`, `LoadTunnelState`.
- **No network**, **no real cloudflared binary** when Runner is injected.
- Sandboxed `HOME` so `DefaultConfigDir()` / credential resolution never
  touches the developer’s real `~/.cloudflared`.
- Temp ConfigDir via `t.TempDir()` for every leaf (root Setup always sets one
  unless a leaf already set ConfigDir).

## Steps

1. Initialize empty `DecisionPath` for tree logging.
2. Allocate sandboxed `HomeDir` and set process `HOME` for this test.
3. Allocate `ConfigDir` temp root when unset.
4. Grouping SETUPs narrow partial-detach / last-host / dns / missing-host;
   leaves fill AttachSequence and StopSequence.

## Context

- On-disk layout:
  `{ConfigDir}/managed-tunnels/<tunnelNameSafe>/{lock,state.json,config.yml}`.
- Fake runner must soft-succeed list/create/info/route/run and write dummy
  credential files (see root `DOCTEST.md` `fakeCloudflared` — same pattern as
  `tests/cloudflare-tunnel-attach`).
- Out of scope: StartSession rewrite, spl CLI, Cloudflare account prune.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req.DecisionPath == nil {
		req.DecisionPath = []string{}
	}

	// Sandbox HOME before any Attach / ensureTunnelForSession credential path.
	if strings.TrimSpace(req.HomeDir) == "" {
		req.HomeDir = t.TempDir()
	}
	if err := os.MkdirAll(filepath.Join(req.HomeDir, ".cloudflared"), 0o755); err != nil {
		return err
	}
	// Do not t.Setenv("HOME"): doctest leaves always call t.Parallel().
	// ConfigDir is always set; HomeDir is passed to the fake runner.

	if strings.TrimSpace(req.ConfigDir) == "" {
		req.ConfigDir = t.TempDir()
	}

	t.Logf("cloudflare-tunnel-detach: start DecisionPath=%v ConfigDir=%q HomeDir=%q",
		req.DecisionPath, req.ConfigDir, req.HomeDir)
	return nil
}
```
