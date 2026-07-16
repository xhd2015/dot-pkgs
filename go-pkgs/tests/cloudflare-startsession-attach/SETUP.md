# Scenario

**Feature**: StartSession wraps Attach for multi-host managed tunnels

```
# production
StartSession(opts with ConfigDir, Runner, Domain, ...)
  -> Attach(AttachOptions{ Domain, LocalURL, TunnelName, ConfigDir, Log, Runner, DNSDeleter })
  -> Session PublicBaseURL + managed Stop

# doctest path (this tree — package API, network free)
Mode=first-session|multi-host|runner
  -> cloudflare.StartSession with fake CommandRunner + SessionOptions.ConfigDir
  -> Response.Session / State / Config / RunCount / Stop
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`
  (local package in this go-pkgs module; no replace).
- **Coverage backfill / GREEN expected:** `SessionOptions.ConfigDir` and
  `StartSession` → managed multi-host `Attach` are present.
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
4. Grouping SETUPs narrow first-session / multi-host / runner; leaves fill
   Domain, LocalURL, TunnelName, Sequence, and StopSequence.

## Context

- On-disk layout:
  `{ConfigDir}/managed-tunnels/<tunnelNameSafe>/{lock,state.json,config.yml}`.
- Fake runner must soft-succeed list/create/info/route/run and write dummy
  credential files (see root `DOCTEST.md` `fakeCloudflared` — same pattern as
  `tests/cloudflare-tunnel-attach` / `tests/cloudflare-tunnel-detach`).
- Out of scope: spl CLI, docs polish.

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

	// Sandbox HOME before any StartSession / ensureTunnelForSession credential path.
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

	t.Logf("cloudflare-startsession-attach: start DecisionPath=%v ConfigDir=%q HomeDir=%q",
		req.DecisionPath, req.ConfigDir, req.HomeDir)
	return nil
}
```
