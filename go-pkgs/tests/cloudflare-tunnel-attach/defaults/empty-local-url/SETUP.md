# Scenario

**Feature**: empty LocalURL defaults to http://127.0.0.1:6321 in Hosts

```
# Attach
LocalURL=""
  -> Hosts[domain].Service = "http://127.0.0.1:6321"
```

## Preconditions

- LocalURL is empty string.
- Domain and TunnelName non-empty.
- Fake runner succeeds.

## Steps

1. Clear LocalURL.
2. Set Domain and TunnelName fixtures.
3. Run single Attach.

## Context

- Same default as StartSession so multi-host Attach matches single-host UX.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "empty-local-url")
	req.Domain = "defaults-url.example.com"
	req.LocalURL = ""
	req.TunnelName = "team-shared"
	return nil
}
```
