# Scenario

**Feature**: Detach of a hostname not present in the registry

```
# missing host
Detach(domain not in Hosts)
  -> no-op success (preferred)
  -> other hosts unchanged
```

## Preconditions

- Package `Detach` API (ViaDetach path).
- Prefer success when domain missing rather than hard error.
- TunnelName default `team-shared` when unset.

## Steps

1. Append `missing-host` to DecisionPath.
2. Set TunnelName when unset.
3. Leaf attaches a real host then Detaches a different missing domain.

## Context

- Requirement scenario 5: Detach missing domain → no-op success preferred.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "missing-host")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
