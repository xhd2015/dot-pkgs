# Scenario

**Feature**: SpaceIndexForWindow succeeds when window’s space is a first-display type-0 Desktop

```
inject ManagedDisplays + WindowSpaceIDs(type0)
  -> SpaceIndexForWindow -> 0-based dense index
```

## Steps

1. Default single-display canonical type-0 spaces unless leaf overrides.
2. Leaves set WindowID and WindowSpaceIDs for the success path under test.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	if len(req.Spaces) == 0 && len(req.Displays) == 0 {
		req.Spaces = canonicalType0Spaces()
	}
	if req.WindowID == 0 {
		req.WindowID = 1001
	}
	return nil
}
```
