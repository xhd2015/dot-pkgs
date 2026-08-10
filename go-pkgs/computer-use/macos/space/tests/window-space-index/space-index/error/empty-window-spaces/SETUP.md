# Scenario

**Feature**: empty CGSCopySpacesForWindows result → detectable error

```
Displays type0 [3,132,234] + CopySpaces([]) empty
  -> SpaceIndexForWindow -> ErrSpaceNotFound
```

## Steps

1. Canonical type-0 displays present.
2. Force empty window space id list (`EmptyWindowSpaces`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Spaces = canonicalType0Spaces()
	req.EmptyWindowSpaces = true
	req.WindowSpaceIDs = nil
	return nil
}
```
