# Scenario

**Feature**: unparseable local or latest → no update

```
NeedsUpdate("", "0.2.0") -> false
# also: garbage local/latest must not force true
```

## Steps

1. Set empty local and valid latest (representative unparseable path).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LocalVer = ""
	req.LatestVer = "0.2.0"
	return nil
}
```
