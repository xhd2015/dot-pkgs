# Scenario

**Feature**: TunnelState save/load against state.json under a tunnel directory

```
# state I/O
SaveTunnelState(dir, st) -> state.json on disk
LoadTunnelState(dir) -> *TunnelState | empty Hosts when missing
```

## Preconditions

- Uses temp dirs only (`t.TempDir()` in Run when StateDir unset).
- No process, flock, or network.

## Steps

1. Append `state` to DecisionPath.
2. Leaves set Mode (`save_load` or `load_missing`) and fixtures.

## Context

- Missing file policy: empty Hosts map + nil error (documented attach-friendly choice).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "state")
	return nil
}
```
