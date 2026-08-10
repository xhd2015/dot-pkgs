## Expected

- `Focus` returns a non-nil error.
- Error message contains the mock Exec error text (`req.ExecError`).

## Errors

- Propagated Exec failure.

## Exit Code

- N/A (library)

```go
import (
	"errors"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	mockErr := errors.New(req.ExecError)
	cfg := &iterm2.FocusConfig{
		Exec: func(script string) error {
			return mockErr
		},
	}
	ferr := iterm2.Focus(sessionRefFromInput(req.FocusRef), cfg)
	if ferr == nil {
		t.Fatal("Focus must propagate Exec error; got nil")
	}
	if !errors.Is(ferr, mockErr) && !strings.Contains(ferr.Error(), req.ExecError) {
		t.Fatalf("Focus error = %v, want to contain/wrap %q", ferr, req.ExecError)
	}
}
```
