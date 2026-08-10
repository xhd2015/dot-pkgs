# Scenario

**Feature**: equal versions do not need update

```
NeedsUpdate("0.147.0", "0.147.0") -> false
```

## Steps

1. Set local and latest to the same semver.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LocalVer = "0.147.0"
	req.LatestVer = "0.147.0"
	return nil
}
```
