# Scenario

**Feature**: after StartSession A+B, Stop(A) leaves only B in managed registry

```
# Sequence
1. StartSession a.example.com -> :7001
2. StartSession b.example.com -> :7002
3. Session.Stop for A
  -> state Hosts only B
  -> B service :7002 retained
```

## Preconditions

- Fresh registry; TunnelName `team-shared` from parent Setup.
- Two StartSession steps then one Stop of A via Session.Stop.
- Sessions must be managed (StartSession → Attach) so Stop detaches one host.

## Steps

1. Set Sequence to A then B with distinct LocalURLs.
2. Set StopSequence to stop A only.
3. Clear ExpectError.

## Context

- Requirement scenario 3 / exit criteria: Stop(A) leaves B.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "stop-leaves-sibling")
	req.Domain = ""
	req.LocalURL = ""
	req.Sequence = []sessionStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:7001"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com"},
	}
	req.ExpectError = false
	return nil
}
```
