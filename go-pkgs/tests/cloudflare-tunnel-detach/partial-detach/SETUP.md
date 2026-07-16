# Scenario

**Feature**: Stop one host while sibling hosts remain attached

```
# partial detach
Attach(A), Attach(B); Stop(A)
  -> Hosts = {B}
  -> config: B host rule + http_status:404
  -> connector still logical-up
```

## Preconditions

- Shared ConfigDir and TunnelName across two Attach calls.
- Same fake runner for Attach and subsequent Stop.
- Domains A and B are distinct.
- Stop uses Session.Stop for the Attach session of A (managed path).

## Steps

1. Append `partial-detach` to DecisionPath.
2. Set default TunnelName when unset.
3. Leaf fills AttachSequence A+B and StopSequence for A.

## Context

- Requirement scenario 1: Attach A+B, Stop(A) → only B remains; connector still up.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "partial-detach")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
