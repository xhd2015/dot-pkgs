## Expected

- `err` is non-nil.
- Error is detectable as `ErrNotUserSpace` or `ErrSpaceNotFound` (not silent success).
- Index is not treated as a valid user Desktop (must not succeed with 0).

## Errors

- Unknown space id outside the type-0 map.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for unknown space id, got nil")
	}
	if !errors.Is(err, space.ErrNotUserSpace) && !errors.Is(err, space.ErrSpaceNotFound) {
		t.Fatalf("err=%v, want errors.Is(..., ErrNotUserSpace|ErrSpaceNotFound)", err)
	}
}
```
