# Scenario

**Feature**: an unknown home still yields the system bundles

```
WellKnownCandidates("") -> 2 system paths
```

## Steps

1. Clear `Home`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Home = ""
	return nil
}
```
