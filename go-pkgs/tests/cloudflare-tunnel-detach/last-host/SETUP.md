# Scenario

**Feature**: releasing the last hostname stops the managed connector

```
# last host
Stop last remaining domain
  -> Hosts empty
  -> ConnectorPID == 0
  -> config has no host rules (404-only OK)
```

## Preconditions

- Managed Attach setup with fake runner.
- TunnelName default `team-shared` when unset.
- Leaves cover single-host Stop and second Stop after partial detach.

## Steps

1. Append `last-host` to DecisionPath.
2. Set TunnelName when unset.
3. Descend into single-host-stops / second-stop-clears.

## Context

- Requirement scenarios 2 and 3: last detach tears down connector.
- Partial-then-last (scenario 2) is a separate leaf from alone Stop (scenario 3).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "last-host")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
