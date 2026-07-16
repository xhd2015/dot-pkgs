# Scenario

**Feature**: ManagedTunnelDir resolves per-tunnel directory under managed-tunnels

```
# ManagedTunnelDir
configDir + tunnelName
  -> Join(ManagedTunnelsRoot(configDir), TunnelNameSafe(tunnelName))
  | error when tunnelName empty
```

## Preconditions

- Mode is `tunnel_dir`.
- Leaves set ConfigDir and TunnelName (empty name for error leaf).

## Steps

1. Set Mode to `tunnel_dir`.
2. Append `tunnel-dir` to DecisionPath.

## Context

- Empty tunnel name must error (do not silently use `"_"` for dir resolution).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "tunnel_dir"
	req.DecisionPath = append(req.DecisionPath, "tunnel-dir")
	return nil
}
```
