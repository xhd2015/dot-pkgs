## Expected

- `err` is nil.
- `resp.Mode` is `color.Auto`.

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
		t.Fatalf("ModeFromFlags(false, false): unexpected error: %v", err)
	}
	if resp.Mode != color.Auto {
		t.Fatalf("ModeFromFlags(false, false): Mode=%v, want Auto", resp.Mode)
	}
}
```
