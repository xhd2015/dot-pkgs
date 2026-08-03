# Scenario

**Feature**: window on last user Desktop (space id 234) → last index

```
Displays type0 [3,132,234] + WindowSpaceIDs[234]
  -> SpaceIndexForWindow -> 2
```

## Steps

1. Inject window space id **234** (last type-0 entry).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.WindowSpaceIDs = []uint64{234}
	return nil
}
```
