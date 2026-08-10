# Scenario

**Feature**: empty names slice returns empty items with nil error

```
LookupPaths([], opts) -> LookupItems{}, nil
```

## Steps

1. Set `Names` to an empty non-nil or nil empty slice (len 0).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{}
	return nil
}
```
