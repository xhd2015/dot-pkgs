# Scenario

**Feature**: team-shared tunnel name resolves under managed-tunnels

```
# ManagedTunnelDir
configDir=/tmp/cf-home, tunnelName=team-shared
  -> /tmp/cf-home/managed-tunnels/team-shared
```

## Preconditions

- Tunnel name is already safe (`team-shared`).
- ConfigDir is a pure path fixture (need not exist).

## Steps

1. Set ConfigDir `/tmp/cf-home` and TunnelName `team-shared`.

## Context

- Requirement scenario 4: full path under managed-tunnels.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "team-shared")
	req.ConfigDir = "/tmp/cf-home"
	req.TunnelName = "team-shared"
	return nil
}
```
