# Scenario

**Feature**: HitTest maps view-local coordinates onto half-open hit rectangles

```
# local (x, localY) against Hit list; first match wins
Hits + (x, localY) -> HitTest -> (Hit, ok)
# y half-open: y0 ≤ localY < y1
```

## Preconditions

- `req.Op = "hit-test"`.
- Hits use the left/run chip row from `leftRunHits()` unless a leaf overrides.

## Steps

1. Set Op to hit-test and install left/run hits.
2. Leaf sets `(X, LocalY)` for hit vs miss at y1.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "hit-test"
	req.Hits = leftRunHits()
	return nil
}
```
