# Scenario

**Feature**: window space id present only on second display → error (first display only)

```
Display0 type0 [3,132,234] + Display1 type0 [400,401]
  + WindowSpaceIDs[400]
  -> SpaceIndexForWindow -> error (400 not in first-display type0 map)
```

## Steps

1. Two-display fixture; window maps to second-display-only id **400**.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Spaces = nil
	req.Displays = []DisplayInput{
		{Spaces: canonicalType0Spaces()},
		{Spaces: []SpaceInfoInput{
			{ID: 400, Type: 0},
			{ID: 401, Type: 0},
		}},
	}
	req.WindowSpaceIDs = []uint64{400}
	return nil
}
```
