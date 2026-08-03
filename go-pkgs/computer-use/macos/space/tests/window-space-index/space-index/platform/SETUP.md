# Scenario

**Feature**: non-darwin platform gate for SpaceIndexForWindow

```
WithPlatformGOOS("linux") -> SpaceIndexForWindow -> ErrUnsupportedPlatform
```

## Steps

1. Leaves set `ForceGOOS` to a non-darwin value.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	// Leaves set ForceGOOS.
	return nil
}
```
