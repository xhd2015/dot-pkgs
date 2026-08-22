# Scenario

**Feature**: hierarchy shape of windows / tabs / sessions in the snapshot

```
PhasedFixture Windows -> Capture -> Snapshot.Windows + Summary counts
```

## Steps

1. Grouping requires iTerm running for all hierarchy leaves.
2. Child leaves set concrete window fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermRunning = true
	return nil
}
```
