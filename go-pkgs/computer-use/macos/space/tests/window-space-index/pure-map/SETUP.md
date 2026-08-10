# Scenario

**Feature**: pure type-0 index build and resolve (no SkyLight / no platform gate)

```
Spaces (id+type) -> BuildUserSpaceIndex / ResolveWindowSpaceIndex
  -> dense map or index | error
```

## Steps

1. Default phase family to pure helpers (leaves set exact Phase).
2. Fixtures use in-memory `SpaceInfoInput` only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaves set Phase to pure-build-index or pure-resolve.
	if req.Phase == "" {
		req.Phase = "pure-build-index"
	}
	return nil
}
```
