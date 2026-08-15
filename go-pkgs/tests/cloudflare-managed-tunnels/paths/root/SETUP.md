# Scenario

**Feature**: ManagedTunnelsRoot joins ConfigDir with fixed segment

```
# ManagedTunnelsRoot
configDir -> Join(configDir, "managed-tunnels")
```

## Preconditions

- Mode is `managed_root`.
- ConfigDir is a plain absolute-style path string (no real dir required).

## Steps

1. Set Mode to `managed_root`.
2. Append `root` to DecisionPath.

## Context

- Segment name is always `managed-tunnels` (not a product name).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = "managed_root"
	req.DecisionPath = append(req.DecisionPath, "root")
	return nil
}
```
