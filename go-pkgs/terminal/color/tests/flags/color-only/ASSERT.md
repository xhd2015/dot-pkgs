## Expected

- `err` is nil.
- `resp.Mode` is `color.Always`.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("ModeFromFlags(true, false): unexpected error: %v", err)
	}
	if resp.Mode != color.Always {
		t.Fatalf("ModeFromFlags(true, false): Mode=%v, want Always", resp.Mode)
	}
}
```
