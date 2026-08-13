## Expected

- `err` is nil.
- `resp.Out` is `""` — no SGR prefix and no reset when text is empty.

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
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Style{Enabled:true}.Green(%q): unexpected error: %v", req.Text, err)
	}
	if resp.Out != "" {
		t.Fatalf("Style{Enabled:true}.Green(\"\"): got %q, want empty string (no escape pair)", resp.Out)
	}
}
```
