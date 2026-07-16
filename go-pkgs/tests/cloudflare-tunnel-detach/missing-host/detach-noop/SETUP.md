# Scenario

**Feature**: Detach unknown domain is a successful no-op; siblings untouched

```
# Sequence
1. Attach b.example.com -> :7002  (registry has only B)
2. Detach a.example.com          (A never attached)
  -> err == nil
  -> Hosts still only B
  -> connector remains up (PID > 0 or prior run count retained)
```

## Preconditions

- Attach B only; Detach A which is not in Hosts.
- ViaDetach=true.
- ExpectError=false (prefer no-op success).

## Steps

1. AttachSequence: B only.
2. StopSequence: Detach A (missing).

## Context

- Requirement scenario 5: prefer no-op success if host not in map.
- Proves Detach does not wipe unrelated hosts when domain is absent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "detach-noop")
	req.AttachSequence = []attachStep{
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com", ViaDetach: true},
	}
	req.ExpectError = false
	req.FailDNS = false
	return nil
}
```
