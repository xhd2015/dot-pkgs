# Scenario

**Feature**: Teardown=true full last-host cleanup vs sibling-safe partial

```
Teardown + last host Stop
  -> DNS delete + tunnel delete -f + managed dir gone + creds gone

Teardown + sibling Stop
  -> DNS delete for detached host only; tunnel/dir kept
```

## Preconditions

- `req.Teardown = true` on leaves.
- Fake runner soft-succeeds `tunnel delete`.
- DNSDeleter wired via harness Attach/Detach.

## Steps

1. Grouping only; leaves set AttachSequence / StopSequence.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "teardown")
	req.Teardown = true
	return nil
}
```
