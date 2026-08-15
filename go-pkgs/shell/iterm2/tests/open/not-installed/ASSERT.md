## Expected

- Returns `ErrNotInstalled`.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if !errors.Is(err, iterm2.ErrNotInstalled) {
		t.Fatalf("error = %v, want ErrNotInstalled", err)
	}
}
```