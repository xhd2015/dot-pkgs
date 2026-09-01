# Scenario

**Feature**: Teardown with siblings only deletes DNS for the detached host

```
Attach A + B; Stop(A) with Teardown
  -> DNS delete for A
  -> TunnelDeleteCount == 0
  -> ManagedDir remains; Hosts only B
```

## Steps

1. Attach two hosts.
2. Stop A only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "siblings-keep-tunnel")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6321"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:6322"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com"},
	}
	return nil
}
```
