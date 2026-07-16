# Scenario

**Feature**: managed-tunnels registry paths, state.json I/O, and pure multi-host ingress build

```
# production
attach host -> LoadTunnelState -> merge Hosts -> SaveTunnelState
            -> BuildConfigFromState -> WriteConfig -> (process)

# doctest path (this tree — package API, network free)
Mode=managed_root|name_safe|tunnel_dir
  -> ManagedTunnelsRoot / TunnelNameSafe / ManagedTunnelDir
  -> Response.Path | error

Mode=save_load|load_missing
  -> SaveTunnelState + LoadTunnelState | LoadTunnelState only
  -> Response.State

Mode=build_config
  -> BuildConfigFromState(StateIn) [x2 for determinism]
  -> Response.Config[, Config2]
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`
  (local package in this go-pkgs module; no replace).
- **Coverage backfill / GREEN expected:** managed path helpers, state I/O, and
  `BuildConfigFromState` are present in `cloudflare/managed.go`.
- Existing helpers still valid: `WriteConfig`, `Config`,
  `IngressRule`, `DefaultConfigDir` — this tree does not start processes.
- **No network**, **no cloudflared binary**, **no flock**, **no DNS**.
- Temp dirs via `t.TempDir()` only (save/load / missing-state leaves).

## Steps

1. Initialize empty `DecisionPath` for tree logging.
2. Grouping SETUPs set `Mode`; leaves fill path fixtures or `StateIn`.
3. `Run` dispatches on `Mode` to the target cloudflare package APIs.

## Context

- Layout: `{ConfigDir}/managed-tunnels/<tunnelNameSafe>/state.json`.
- Load missing prefers empty `Hosts` + nil error (attach ease).
- Ingress order: sorted hostnames, then `http_status:404` catch-all last.
- Out of scope: Attach/Detach, flock, process, DNS, spl CLI, StartSession.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.DecisionPath == nil {
		req.DecisionPath = []string{}
	}
	t.Logf("cloudflare-managed-tunnels: start DecisionPath=%v Mode=%q", req.DecisionPath, req.Mode)
	return nil
}
```
