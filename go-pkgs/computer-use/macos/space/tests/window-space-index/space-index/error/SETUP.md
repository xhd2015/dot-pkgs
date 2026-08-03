# Scenario

**Feature**: SpaceIndexForWindow returns a detectable error (never silent index 0)

```
inject fixture where window space is missing / non-type0 / other display
  -> SpaceIndexForWindow -> ErrNotUserSpace | ErrSpaceNotFound
```

## Steps

1. Leaves configure the failing fixture (empty CopySpaces, non-user id, or 2nd-display-only id).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	if req.WindowID == 0 {
		req.WindowID = 1001
	}
	return nil
}
```
