# Scenario

**Feature**: empty optional Attach fields resolve to package defaults

```
# defaults
TunnelName="" -> DefaultTunnelName
LocalURL=""   -> http://127.0.0.1:6321
```

## Preconditions

- Domain is non-empty (defaults only apply after Domain validates).
- Fake runner path (lifecycle-style Attach success).
- ConfigDir already sandboxed by root Setup.

## Steps

1. Append `defaults` to DecisionPath.
2. Leaves clear TunnelName or LocalURL and attach one host.

## Context

- Mirrors StartSession defaults for LocalURL and TunnelName so Attach is drop-in
  for multi-host callers.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "defaults")
	// Stable domain for defaults leaves unless a leaf overrides.
	if req.Domain == "" {
		req.Domain = "defaults.example.com"
	}
	return nil
}
```
