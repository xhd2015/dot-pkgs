# Scenario

**Feature**: BuildConfigFromState emits multi-host ingress with sorted order and 404 last

```
# pure build
TunnelState{Hosts}
  -> BuildConfigFromState
  -> Config.Tunnel / CredentialsFile from state
  -> Ingress: sorted host rules + http_status:404 catch-all (empty hostname)
```

## Preconditions

- Mode is `build_config` for all leaves.
- StateIn.Hosts non-nil (may be 1 or 2 hosts).
- No filesystem or process.

## Steps

1. Set Mode to `build_config`.
2. Append `build-config` to DecisionPath.
3. Leaves populate StateIn hosts and metadata.

## Context

- Catch-all must be last; hostnames sorted for determinism (requirement 8–11).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = "build_config"
	req.DecisionPath = append(req.DecisionPath, "build-config")
	return nil
}
```
