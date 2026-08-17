# Scenario

**Feature**: PinPatch maps a go version spelling to a pinned SDK name

```
# known major.minor from kool table; naked 1.19 same as go1.19
caller goVersion -> PinPatch -> go1.Y.Z

# already-full patch is identity; unknown major.minor keeps go prefix
go1.19.13 / 1.19.13 -> go1.19.13
go1.99 -> go1.99
```

## Steps

1. Set `req.Op` to `pin`.
2. Leaf `Setup` sets `PinInputs` or `GoVersion`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "pin"
	return nil
}
```
