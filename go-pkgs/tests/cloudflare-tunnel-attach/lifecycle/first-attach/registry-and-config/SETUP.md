# Scenario

**Feature**: first attach writes one host into state and host+404 config under managed dir

```
# Attach empty registry
Attach(a.example.com -> :6321)
  -> state.Hosts len=1
  -> config.yml exists under ManagedTunnelDir
  -> ingress: host rule then http_status:404
```

## Preconditions

- Domain `a.example.com`, LocalURL `http://127.0.0.1:6321`, TunnelName
  `team-shared` (from lifecycle Setup).
- Fresh ConfigDir (no prior state).

## Steps

1. Use first-attach fixtures (no Sequence).
2. Run Attach once via harness.

## Context

- Requirement scenarios 1 and 5: first attach registry + config path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "registry-and-config")
	// Pin fixtures explicitly so this leaf is self-describing.
	req.Domain = "a.example.com"
	req.LocalURL = "http://127.0.0.1:6321"
	req.Sequence = nil
	return nil
}
```
