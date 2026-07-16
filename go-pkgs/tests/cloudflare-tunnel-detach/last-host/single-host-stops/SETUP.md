# Scenario

**Feature**: Attach one host then Stop it → empty registry and connector down

```
# Sequence
1. Attach a.example.com -> :6321
2. Session.Stop(A)
  -> Hosts empty
  -> ConnectorPID == 0
  -> no hostname ingress rules (404-only OK)
```

## Preconditions

- Fresh registry; single Attach then Stop via Session.Stop.
- TunnelName `team-shared` from parent.

## Steps

1. AttachSequence: one host A.
2. StopSequence: Stop A (ViaDetach=false).

## Context

- Requirement scenario 3: Attach A, Stop(A) alone → empty hosts; connector down.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "single-host-stops")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6321"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com", ViaDetach: false},
	}
	req.ExpectError = false
	req.FailDNS = false
	return nil
}
```
