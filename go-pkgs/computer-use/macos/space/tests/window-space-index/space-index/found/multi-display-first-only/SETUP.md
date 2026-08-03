# Scenario

**Feature**: multi-display fixture still indexes only first display’s type-0 list

```
Display0 type0 [3,132,234] + Display1 type0 [400,401]
  + WindowSpaceIDs[132]
  -> SpaceIndexForWindow -> 1
  # 400/401 must not shift first-display indices
```

## Steps

1. Two-display ManagedDisplaySpaces fixture.
2. Window maps to first-display id **132**.

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
	req.WindowSpaceIDs = []uint64{132}
	return nil
}
```
