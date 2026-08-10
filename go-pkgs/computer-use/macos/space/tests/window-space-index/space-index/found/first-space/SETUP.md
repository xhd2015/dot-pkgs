# Scenario

**Feature**: window on first user Desktop (space id 3) → index 0

```
Displays type0 [3,132,234] + WindowSpaceIDs[3]
  -> SpaceIndexForWindow -> 0
```

## Steps

1. Inject window space id **3** (first type-0 entry).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WindowSpaceIDs = []uint64{3}
	return nil
}
```
