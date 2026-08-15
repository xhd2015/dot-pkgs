# Scenario

**Feature**: StartSession with fake Runner calls tunnel run and uses managed config

```
# Sequence
1. StartSession a.example.com -> :6321, Runner=fake, ConfigDir=tmp
  -> err == nil (no panic)
  -> RunCount ≥ 1
  -> run --config points at managed config.yml when --config present
```

## Preconditions

- Fresh ConfigDir; single StartSession.
- Domain `a.example.com`, LocalURL `http://127.0.0.1:6321`.

## Steps

1. Pin Domain / LocalURL fixtures.
2. Clear Sequence / StopSequence.
3. Run StartSession once via harness.

## Context

- Requirement scenario 4.
- Hard assert: RunCount ≥ 1 and managed Hosts/config path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "exec-run")
	req.Domain = "a.example.com"
	req.LocalURL = "http://127.0.0.1:6321"
	req.Sequence = nil
	req.StopSequence = nil
	req.ExpectError = false
	return nil
}
```
