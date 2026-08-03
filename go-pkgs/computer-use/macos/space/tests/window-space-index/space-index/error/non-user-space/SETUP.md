# Scenario

**Feature**: window on a non-type-0 / unmapped space id → detectable error (not 0)

```
Displays type0 [3,132,234] + type4 id 50
  + WindowSpaceIDs[50]
  -> SpaceIndexForWindow -> ErrNotUserSpace (not index 0)
```

## Steps

1. Spaces include a type-4 entry that is **not** in the user map.
2. Window maps to that type-4 id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Spaces = []SpaceInfoInput{
		{ID: 3, Type: 0},
		{ID: 132, Type: 0},
		{ID: 234, Type: 0},
		{ID: 50, Type: 4},
	}
	req.WindowSpaceIDs = []uint64{50}
	return nil
}
```
