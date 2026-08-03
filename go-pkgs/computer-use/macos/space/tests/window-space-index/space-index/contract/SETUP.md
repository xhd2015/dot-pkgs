# Scenario

**Feature**: production CopySpaces path uses mask 7 (observable via injectable hook)

```
WithCopySpacesForWindows spy <- SpaceIndexForWindow
  -> captured mask == 7
```

## Steps

1. Leaves enable `CaptureMask` and return a valid type-0 space id from the spy.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.CaptureMask = true
	return nil
}
```
