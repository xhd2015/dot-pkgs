# Scenario

**Feature**: last-host Teardown removes DNS, tunnel, creds, and managed dir

```
Attach a.example.com; Session.Stop
  -> DNSDeleteCount >= 1
  -> TunnelDeleteCount >= 1
  -> ManagedDirRemoved
  -> CredFileRemoved
```

## Steps

1. Attach single host with Teardown (from parent).
2. Stop via Session.Stop.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "last-host-full")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6321"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com"},
	}
	return nil
}
```
