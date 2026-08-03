## Expected

- `err` is non-nil.
- Prefer `errors.Is(err, space.ErrSpaceNotFound)`; `ErrNotUserSpace` also acceptable if implementer unifies.
- Must not succeed with index 0.

## Errors

- No space ids returned for the window.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err == nil {
		idx := -1
		if resp != nil {
			idx = resp.Index
		}
		t.Fatalf("expected error for empty window spaces, got nil (Index=%d)", idx)
	}
	if !errors.Is(err, space.ErrSpaceNotFound) && !errors.Is(err, space.ErrNotUserSpace) {
		t.Fatalf("err=%v, want errors.Is(..., ErrSpaceNotFound|ErrNotUserSpace)", err)
	}
}
```
