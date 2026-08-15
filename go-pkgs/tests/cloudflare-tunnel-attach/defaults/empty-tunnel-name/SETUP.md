# Scenario

**Feature**: empty TunnelName resolves to DefaultTunnelName

```
# Attach
TunnelName=""
  -> registry under ManagedTunnelDir(configDir, DefaultTunnelName)
  -> Session.TunnelName == DefaultTunnelName
```

## Preconditions

- TunnelName is empty string.
- Domain is non-empty; LocalURL explicit.
- Fake runner succeeds ensure + route + run.

## Steps

1. Clear TunnelName.
2. Set Domain and LocalURL fixtures.
3. Run single Attach via shared harness.

## Context

- Managed dir segment must use TunnelNameSafe(DefaultTunnelName), not an empty
  segment.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "empty-tunnel-name")
	req.TunnelName = ""
	req.Domain = "defaults-tunnel.example.com"
	req.LocalURL = "http://127.0.0.1:7001"
	return nil
}
```
