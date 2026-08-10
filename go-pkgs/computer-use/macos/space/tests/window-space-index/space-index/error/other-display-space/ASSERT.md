## Expected

- `err` is non-nil (`ErrNotUserSpace` or `ErrSpaceNotFound`).
- Must not return a second-display index (e.g. 0 for 400 on display1).

## Errors

- Space id is not on the first display’s type-0 list.

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
		t.Fatalf("expected error for other-display space id, got nil (Index=%d)", idx)
	}
	if !errors.Is(err, space.ErrNotUserSpace) && !errors.Is(err, space.ErrSpaceNotFound) {
		t.Fatalf("err=%v, want errors.Is(..., ErrNotUserSpace|ErrSpaceNotFound)", err)
	}
}
```
