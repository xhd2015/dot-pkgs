# Scenario

**Feature**: two StartSession calls share managed tunnel registry

```
# multi-host
StartSession(A); StartSession(B) same TunnelName + ConfigDir
  -> Hosts has A and B
Stop(A)
  -> only B remains
```

## Preconditions

- Domain non-empty per step.
- ConfigDir and HOME already sandboxed by root Setup.
- TunnelName default for this branch: `team-shared` unless a leaf overrides.
- Shared fake runner across Sequence steps.

## Steps

1. Append `multi-host` to DecisionPath.
2. Set a stable TunnelName when unset.
3. Descend into hosts-merge / stop-leaves-sibling.

## Context

- Requirement scenarios 2 and 3 / exit criteria:
  StartSession(A) then StartSession(B same tunnel) → state has two hosts;
  Stop(A) leaves B.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "multi-host")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
