# Scenario

**Feature**: Resolve maps absolute mouse coordinates onto a view-local Hit

```
# known: localY = absY - *OriginY
# dual: try top (absY) then bottom (absY - BottomOriginY(height, viewLines))
Abs + OriginY? + Hits -> Resolve -> {OK, Hit, LocalY, Kind}
```

## Preconditions

- `req.Op = "resolve"`.
- Terminal height and viewLines are set by sub-branches for mid-pane or dual fixtures.

## Steps

1. Set Op to resolve.
2. known-mid sets OriginY; dual-* leave OriginY nil.
3. Leaves set AbsX/AbsY and assert Hit ID + Kind.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "resolve"
	return nil
}
```
