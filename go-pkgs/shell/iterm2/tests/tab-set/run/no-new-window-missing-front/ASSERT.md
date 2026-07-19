## Expected

- `RunTabSet` returns an error.
- Error is or wraps `ErrNoITermWindow`.

## Errors

- `ErrNoITermWindow` when NoNewWindow cannot target a window.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	cfg := buildRunConfig(req, &scripts)
	_, rerr := iterm2.RunTabSet(tabSetSpecFromReq(req), iterm2.TabSetRunOptions{
		Mode: iterm2.TabSetRunNoNewWindow,
	}, cfg)
	if rerr == nil {
		t.Fatal("expected ErrNoITermWindow, got nil")
	}
	if !errors.Is(rerr, iterm2.ErrNoITermWindow) {
		t.Fatalf("error = %v, want errors.Is(..., ErrNoITermWindow)", rerr)
	}
}
```
