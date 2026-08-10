# Scenario

**Feature**: marshal/unmarshal preserves all Event fields including host and payload

```
# full fixture with host set
fixture Event (id, ts, source, type, host, payload) -> round-trip equal
```

## Steps

1. Keep the root fixture Event (all fields populated, including `Host`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// req.Event is fixtureEvent() from root Setup (Host = "test-host")
	if req.Event.Host == "" {
		t.Fatal("fixture Event must include Host for this leaf")
	}
	return nil
}
```
