# Scenario

**Feature**: second distinct hostname attaches to the same managed tunnel

```
# sequence
Attach(A) then Attach(B) same TunnelName + ConfigDir
  -> Hosts {A,B}
  -> config both hostnames + 404 last
  -> run count increases (restart / second run)
```

## Preconditions

- Shared ConfigDir and TunnelName across two Attach calls.
- Same fake runner instance for the whole Sequence (harness-owned).
- Domains A and B are distinct.

## Steps

1. Append `second-host` to DecisionPath.
2. Leaves populate Sequence with two attach steps.

## Context

- Requirement scenario 2: second host same tunnel → merge + restart.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "second-host")
	return nil
}
```
