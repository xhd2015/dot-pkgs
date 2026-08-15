# Scenario

**Feature**: after partial Stop(A), Stop(B) clears registry and stops connector

```
# Sequence
1. Attach a.example.com -> :7001
2. Attach b.example.com -> :7002
3. Session.Stop(A)
4. Session.Stop(B)
  -> Hosts empty
  -> ConnectorPID == 0
  -> no hostname ingress rules
```

## Preconditions

- Full A+B attach then ordered Stop A then Stop B via Session.Stop.
- Complements `partial-detach/siblings-remain` (which stops only at A).

## Steps

1. AttachSequence A then B.
2. StopSequence A then B (ViaDetach=false both).

## Context

- Requirement scenario 2 / exit criteria second half: then Stop(B) → empty +
  connector stopped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "second-stop-clears")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:7001"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com", ViaDetach: false},
		{Domain: "b.example.com", ViaDetach: false},
	}
	req.ExpectError = false
	req.FailDNS = false
	return nil
}
```
