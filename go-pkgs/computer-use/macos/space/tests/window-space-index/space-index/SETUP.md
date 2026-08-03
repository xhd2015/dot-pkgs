# Scenario

**Feature**: `SpaceIndexForWindow` with injectable ManagedDisplaySpaces + CopySpaces fixtures

```
windowID + inject(ManagedDisplays, WindowSpaceIDs|CopySpaces hook, GOOS)
  -> SpaceIndexForWindow -> index | error
```

## Steps

1. Set `req.Phase` to `space-index-for-window`.
2. Leaves supply displays, window id, and injected space ids (or empty / spy).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Phase = "space-index-for-window"
	return nil
}
```
