# Scenario

**Feature**: window on middle user Desktop (space id 132) → index 1

```
Displays type0 [3,132,234] + WindowSpaceIDs[132]
  -> SpaceIndexForWindow -> 1
```

## Steps

1. Inject window space id **132** (requirement golden case).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WindowSpaceIDs = []uint64{132}
	return nil
}
```
