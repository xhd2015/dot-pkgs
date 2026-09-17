# Scenario

**Feature**: both the system and the per-user bundles are searchable

```
WellKnownCandidates(home) -> 4 paths
```

## Steps

1. `Home` is the fixture home from root Setup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	return nil
}
```
