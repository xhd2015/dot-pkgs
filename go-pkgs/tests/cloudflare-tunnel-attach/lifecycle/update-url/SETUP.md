# Scenario

**Feature**: re-attach same hostname with a new LocalURL updates service and restarts

```
# sequence
Attach(A, :6321) then Attach(A, :6322)
  -> Hosts[A].Service = :6322
  -> run count increases on ingress change
```

## Preconditions

- Same Domain twice; LocalURL changes.
- Shared ConfigDir, TunnelName, and fake runner.

## Steps

1. Append `update-url` to DecisionPath.
2. Leaf populates Sequence with two steps for the same host.

## Context

- Requirement scenario 3: update same host LocalURL.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "update-url")
	return nil
}
```
