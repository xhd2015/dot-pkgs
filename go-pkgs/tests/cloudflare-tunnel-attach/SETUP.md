# Scenario

**Feature**: Attach hostname to managed named tunnel (registry + DNS + config + connector)

```
# production
Attach(domain, localURL, tunnelName, configDir, runner)
  -> flock managed dir
  -> ensureTunnelForSession + RouteDNS
  -> merge Hosts[domain] in state.json
  -> BuildConfigFromState -> config.yml
  -> runner Exec tunnel run (fake) when first / ingress changed
  -> Session for domain

# doctest path (this tree — package API, network free)
Mode=validation|defaults|lifecycle
  -> cloudflare.Attach with fake CommandRunner
  -> Response.Session / State / Config / RunCount / RouteDNSCount
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`
  (local package in this go-pkgs module; no replace).
- **Coverage backfill / GREEN expected:** `Attach` / `AttachOptions` present.
- Managed helpers for post-condition reads:
  `ManagedTunnelDir`, `LoadTunnelState`, `BuildConfigFromState` (optional),
  `WriteConfig` (production side).
- **No network**, **no real cloudflared binary** when Runner is injected.
- Sandboxed `HOME` so `DefaultConfigDir()` / credential resolution never
  touches the developer’s real `~/.cloudflared`.
- Temp ConfigDir via `t.TempDir()` for every leaf (root Setup always sets one
  unless a leaf already set ConfigDir).

## Steps

1. Initialize empty `DecisionPath` for tree logging.
2. Allocate sandboxed `HomeDir` and set process `HOME` for this test.
3. Allocate `ConfigDir` temp root when unset.
4. Grouping SETUPs narrow validation / defaults / lifecycle; leaves fill
   Domain, LocalURL, TunnelName, or Sequence.

## Context

- On-disk layout:
  `{ConfigDir}/managed-tunnels/<tunnelNameSafe>/{lock,state.json,config.yml}`.
- Fake runner must soft-succeed list/create/info/route/run and write dummy
  credential files (see root `DOCTEST.md` `fakeCloudflared`).
- Out of scope: Detach/Stop, StartSession, spl CLI, DNS delete.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
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
	t.Setenv("HOME", req.HomeDir)

	if strings.TrimSpace(req.ConfigDir) == "" {
		req.ConfigDir = t.TempDir()
	}

	t.Logf("cloudflare-tunnel-attach: start DecisionPath=%v ConfigDir=%q HomeDir=%q",
		req.DecisionPath, req.ConfigDir, req.HomeDir)
	return nil
}
```
