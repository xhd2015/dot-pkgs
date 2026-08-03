# Scenario

**Feature**: injectable CopySpacesForWindows receives mask **7**

```
WithCopySpacesForWindows(spy) + ManagedDisplays + window
  -> SpaceIndexForWindow
  -> spy mask == 7; resolves index from spy return
```

## Steps

1. CaptureMask enabled (grouping).
2. Canonical type-0 displays; spy returns space id **132**.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Spaces = canonicalType0Spaces()
	req.WindowID = 4242
	req.WindowSpaceIDs = []uint64{132}
	req.CaptureMask = true
	return nil
}
```
