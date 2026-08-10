# Scenario

**Feature**: local older than latest needs update

```
NeedsUpdate("0.1.0", "0.2.0") -> true
```

## Steps

1. Set local `0.1.0`, latest `0.2.0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LocalVer = "0.1.0"
	req.LatestVer = "0.2.0"
	return nil
}
```
