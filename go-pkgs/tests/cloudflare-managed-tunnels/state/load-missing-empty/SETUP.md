# Scenario

**Feature**: LoadTunnelState on missing state.json returns empty Hosts without error

```
# load missing
dir without state.json
  -> LoadTunnelState(dir)
  -> Hosts non-nil empty map, err == nil
```

## Preconditions

- Mode `load_missing`.
- No state.json under the temp path (Run creates empty path).

## Steps

1. Set Mode to `load_missing`.

## Context

- Requirement scenario 7: prefer empty+nil over ErrNotExist for attach ease.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "load-missing-empty")
	req.Mode = "load_missing"
	// StateDir left empty so Run uses a fresh temp path without state.json.
	return nil
}
```
