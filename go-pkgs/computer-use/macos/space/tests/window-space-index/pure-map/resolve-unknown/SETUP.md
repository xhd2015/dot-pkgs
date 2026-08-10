# Scenario

**Feature**: ResolveWindowSpaceIndex errors when space id is not in the type-0 map

```
index{3:0,132:1,234:2} + windowSpaceIDs[999]
  -> ResolveWindowSpaceIndex -> ErrNotUserSpace (or equivalent detectable error)
```

## Steps

1. Phase `pure-resolve`.
2. Canonical type-0 list; resolve unknown id **999**.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "pure-resolve"
	req.Spaces = canonicalType0Spaces()
	req.ResolveSpaceIDs = []uint64{999}
	return nil
}
```
