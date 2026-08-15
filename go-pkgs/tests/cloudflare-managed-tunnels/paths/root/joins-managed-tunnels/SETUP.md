# Scenario

**Feature**: ManagedTunnelsRoot("/tmp/cf") ends with managed-tunnels segment

```
# ManagedTunnelsRoot
configDir=/tmp/cf
  -> /tmp/cf/managed-tunnels
```

## Preconditions

- ConfigDir is the locked requirement fixture `/tmp/cf`.
- No need for `/tmp/cf` to exist on disk (pure join).

## Steps

1. Set ConfigDir to `/tmp/cf`.

## Context

- Requirement scenario 1: exact join result.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "joins-managed-tunnels")
	req.ConfigDir = "/tmp/cf"
	return nil
}
```
