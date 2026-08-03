# Scenario

**Feature**: non-darwin → ErrUnsupportedPlatform

```
WithPlatformGOOS("linux") -> SpaceIndexForWindow -> ErrUnsupportedPlatform
```

## Steps

1. Force platform to `linux` via injectable option.
2. Minimal window id (fixtures optional; platform gate runs first).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.ForceGOOS = "linux"
	req.WindowID = 1
	// Fixtures present so implementer must still gate before using them.
	req.Spaces = canonicalType0Spaces()
	req.WindowSpaceIDs = []uint64{3}
	return nil
}
```
