# Scenario

**Feature**: type!=0 Spaces entries are skipped; indices stay dense over type-0 only

```
Spaces [3 type0, 50 type4, 132 type0, 234 type0]
  + WindowSpaceIDs[132]
  -> SpaceIndexForWindow -> 1
  # not 2 (type4 must not consume an index)
```

## Steps

1. Override Spaces with a mixed-type list.
2. Window maps to **132**.

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
		{ID: 50, Type: 4},
		{ID: 132, Type: 0},
		{ID: 234, Type: 0},
	}
	req.WindowSpaceIDs = []uint64{132}
	return nil
}
```
