# Scenario

**Feature**: empty Domain is rejected by Attach

```
# Attach
Domain=""
  -> error (domain is required)
```

## Preconditions

- Domain is empty string.
- LocalURL / TunnelName / ConfigDir may be set; validation must still fail on Domain.
- ExpectError is true.

## Steps

1. Set Domain to empty.
2. Mark ExpectError so Assert requires non-nil err.
3. Optional LocalURL and TunnelName are intentionally non-empty to prove Domain is the gate.

## Context

- Requirement scenario 4: Domain required; no state file required on failure.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "domain-empty-errors")
	req.Domain = ""
	req.LocalURL = "http://127.0.0.1:6321"
	req.TunnelName = "team-shared"
	req.ExpectError = true
	return nil
}
```
