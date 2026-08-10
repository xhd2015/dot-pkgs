# Scenario

**Feature**: local newer than latest does not need update

```
NeedsUpdate("0.2.0", "0.1.0") -> false
```

## Steps

1. Set local greater than latest.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LocalVer = "0.2.0"
	req.LatestVer = "0.1.0"
	return nil
}
```
