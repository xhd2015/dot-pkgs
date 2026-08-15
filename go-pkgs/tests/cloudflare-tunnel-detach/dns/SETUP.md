# Scenario

**Feature**: best-effort DNS delete on Detach / Stop

```
# DNS path
Detach/Stop(domain) with DNSDeleter
  -> DeleteHostname attempted
  -> DNS error does not fail Detach/Stop
  -> host still removed from registry
```

## Preconditions

- Injectable `DNSDeleter` via `DetachOptions` (fake in harness).
- Fake CommandRunner for Attach setup.
- TunnelName default `team-shared` when unset.

## Steps

1. Append `dns` to DecisionPath.
2. Set TunnelName when unset.
3. Leaf configures FailDNS and Detach path.

## Context

- Requirement scenario 4: DNS delete failure does not fail Stop/Detach.
- Leaf uses package `Detach` with DNSDeleter so DNS inject does not depend on
  AttachOptions growing a DNSDeleter field (Session.Stop → Detach for product path).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "dns")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
