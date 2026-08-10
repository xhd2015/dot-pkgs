## Expected

- `err` is non-nil.
- `errors.Is(err, space.ErrNotUserSpace)` (preferred) or `ErrSpaceNotFound`.
- Must **not** succeed with index 0 (silent fallback forbidden).

## Errors

- Window space id is not a first-display type-0 user Desktop.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		idx := -1
		if resp != nil {
			idx = resp.Index
		}
		t.Fatalf("expected error for non-user space, got nil (Index=%d)", idx)
	}
	if !errors.Is(err, space.ErrNotUserSpace) && !errors.Is(err, space.ErrSpaceNotFound) {
		t.Fatalf("err=%v, want errors.Is(..., ErrNotUserSpace|ErrSpaceNotFound)", err)
	}
}
```
