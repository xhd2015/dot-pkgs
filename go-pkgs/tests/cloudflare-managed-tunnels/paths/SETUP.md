# Scenario

**Feature**: pure path and name helpers for managed-tunnels layout

```
# path surfaces
configDir -> ManagedTunnelsRoot
tunnelName -> TunnelNameSafe
configDir + tunnelName -> ManagedTunnelDir | error
```

## Preconditions

- Mode will be set by child grouping nodes (`managed_root`, `name_safe`,
  `tunnel_dir`).
- No filesystem I/O required for these pure join / sanitize helpers.

## Steps

1. Append `paths` to DecisionPath.
2. Descend into root / name-safe / tunnel-dir branches.

## Context

- Highest significance under this tree: which path helper is under test.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "paths")
	return nil
}
```
