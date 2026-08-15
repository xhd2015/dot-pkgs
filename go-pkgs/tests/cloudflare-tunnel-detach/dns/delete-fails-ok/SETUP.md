# Scenario

**Feature**: DNS delete error is best-effort; Detach still succeeds and removes host

```
# Sequence
1. Attach a.example.com -> :6321
2. Detach(A) with DNSDeleter that returns error
  -> Detach err == nil
  -> Hosts empty (host removed despite DNS failure)
  -> DNSDeleteCount >= 1 for A
```

## Preconditions

- FailDNS=true on harness fakeDNSDeleter.
- Stop step uses ViaDetach=true so DetachOptions.DNSDeleter is injected.
- ExpectError=false (DNS failure must not surface as API error).

## Steps

1. Attach A.
2. Detach A with failing DNS deleter.

## Context

- Requirement scenario 4: DNS delete fails → Stop/Detach returns nil; warning path OK.
- Logging of warning is soft (not asserted); API success + host removal + attempt count are hard.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "delete-fails-ok")
	req.AttachSequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6321"},
	}
	req.StopSequence = []stopStep{
		{Domain: "a.example.com", ViaDetach: true},
	}
	req.FailDNS = true
	req.ExpectError = false
	return nil
}
```
