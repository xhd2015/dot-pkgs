## Expected

- `err` is non-nil.
- `errors.Is(err, space.ErrUnsupportedPlatform)`.

## Errors

- Unsupported platform (macOS only), matching Create/List/CountDesktops.

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
	_ = resp
	if err == nil {
		t.Fatal("expected ErrUnsupportedPlatform, got nil")
	}
	if !errors.Is(err, space.ErrUnsupportedPlatform) {
		t.Fatalf("err=%v, want errors.Is(..., ErrUnsupportedPlatform)", err)
	}
}
```
