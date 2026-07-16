# Scenario

**Feature**: empty tunnel name is rejected when resolving ManagedTunnelDir

```
# ManagedTunnelDir
tunnelName="" -> error
```

## Preconditions

- ConfigDir may be any non-empty path fixture.
- TunnelName is empty string.

## Steps

1. Set ConfigDir and empty TunnelName.
2. Mark ExpectError so Assert requires non-nil err.

## Context

- Requirement scenario 5: empty name → error (prefer over `"_"` fallback).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "empty-name-errors")
	req.ConfigDir = "/tmp/cf-home"
	req.TunnelName = ""
	req.ExpectError = true
	return nil
}
```
