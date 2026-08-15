# Scenario

**Feature**: after A+B, Stop(A) leaves only B and keeps connector up

```
# Sequence
1. Attach a.example.com -> :7001
2. Attach b.example.com -> :7002
3. Session.Stop for A
  -> state Hosts only B
  -> config only B + 404 last
  -> ConnectorPID > 0 OR RunCount increased after Stop (restart with remaining)
```

## Preconditions

- Fresh registry; TunnelName `team-shared` from parent Setup.
- Two Attach steps then one Stop of A via Session.Stop (ViaDetach=false).

## Steps

1. Set AttachSequence to A then B with distinct LocalURLs.
2. Set StopSequence to stop A only.
3. Clear ExpectError.

## Context

- Requirement scenario 1 / exit criteria first half.
- "Connector still up": `state.ConnectorPID > 0` **or** `RunCount > RunCountAfterAttach`
  (restart invoked with remaining hosts).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "siblings-remain")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:7001"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com", ViaDetach: false},
	}
	req.ExpectError = false
	req.FailDNS = false
	return nil
}
```
