## Expected

- `err` is nil.
- `resp.Enabled` is true: Always forces color on a non-TTY.

## Side Effects

- None. Resolve is a pure function of its arguments.

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
		t.Fatalf("Resolve(Always, false, %q): unexpected error: %v", req.NoColorEnv, err)
	}
	if !resp.Enabled {
		t.Fatal("Resolve(Always, false, \"\"): Enabled=false, want true (Always ignores TTY)")
	}
}
```
