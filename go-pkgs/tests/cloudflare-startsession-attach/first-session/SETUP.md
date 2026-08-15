# Scenario

**Feature**: single StartSession with managed ConfigDir preserves PublicBaseURL

```
# first session
StartSession(A, ConfigDir)
  -> PublicBaseURL = https://A
  -> managed Hosts[A] under ManagedTunnelDir
```

## Preconditions

- Domain non-empty.
- ConfigDir and HOME already sandboxed by root Setup.
- TunnelName default for this branch: `team-shared` unless a leaf overrides.
- Fake runner soft-succeeds list/create/info/route/run.

## Steps

1. Append `first-session` to DecisionPath.
2. Set a stable TunnelName when unset.
3. Descend into public-base-url.

## Context

- Requirement scenario 1: SessionOptions.ConfigDir + PublicBaseURL.
- GREEN: StartSession writes managed registry via Attach.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "first-session")
	if req.TunnelName == "" {
		req.TunnelName = "team-shared"
	}
	return nil
}
```
