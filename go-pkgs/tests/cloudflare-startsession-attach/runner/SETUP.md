# Scenario

**Feature**: StartSession works with injectable fake CommandRunner

```
# runner
StartSession(..., Runner=fake)
  -> no panic / err nil
  -> tunnel … run Exec at least once
```

## Preconditions

- Domain non-empty.
- ConfigDir and HOME already sandboxed by root Setup.
- TunnelName default for this branch: `team-shared` unless a leaf overrides.
- Same fake runner soft contracts as attach/detach trees.

## Steps

1. Append `runner` to DecisionPath.
2. Set a stable TunnelName when unset.
3. Descend into exec-run.

## Context

- Requirement scenario 4: fake runner path like Attach tests.
- Also assert managed config/Hosts so leaf proves StartSession wraps Attach.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "runner")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
