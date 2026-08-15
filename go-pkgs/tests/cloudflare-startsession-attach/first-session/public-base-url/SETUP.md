# Scenario

**Feature**: StartSession(A) with ConfigDir yields PublicBaseURL and managed host A

```
# Sequence
1. StartSession a.example.com -> :6321, ConfigDir=tmp, TunnelName=team-shared
  -> Session.PublicBaseURL = https://a.example.com
  -> managed state Hosts[a.example.com] service :6321
  -> Session.ConfigPath under ManagedTunnelDir(ConfigDir, team-shared)
```

## Preconditions

- Fresh ConfigDir (no prior managed state).
- Single StartSession (Sequence empty).
- Domain `a.example.com`, LocalURL `http://127.0.0.1:6321`.

## Steps

1. Pin Domain / LocalURL fixtures.
2. Clear Sequence / StopSequence.
3. Run StartSession once via harness with SessionOptions.ConfigDir.

## Context

- Requirement scenario 1.
- Asserts managed path (StartSession → Attach).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "public-base-url")
	req.Domain = "a.example.com"
	req.LocalURL = "http://127.0.0.1:6321"
	req.Sequence = nil
	req.StopSequence = nil
	req.ExpectError = false
	return nil
}
```
