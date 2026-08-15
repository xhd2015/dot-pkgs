# Scenario

**Feature**: successful Attach lifecycle with fake CommandRunner (no network)

```
# lifecycle
Attach(domain, localURL, tunnelName, configDir, fakeRunner)
  -> ensureTunnel + RouteDNS + merge Hosts + write config + Exec run
  -> Response.State / Config / RunCount / RouteDNSCount
```

## Preconditions

- Domain non-empty for all lifecycle leaves.
- TunnelName default for this branch: `team-shared` unless a leaf overrides.
- ConfigDir and HOME already sandboxed by root Setup.
- Fake runner soft-succeeds list/create/info/route/run and writes dummy creds.

## Steps

1. Append `lifecycle` to DecisionPath.
2. Set a stable TunnelName when unset.
3. Descend into first-attach / second-host / update-url.

## Context

- Exit criteria: first attach → one run; second host → two hosts +
  restart; same host new URL → update + run again; no real network.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "lifecycle")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
