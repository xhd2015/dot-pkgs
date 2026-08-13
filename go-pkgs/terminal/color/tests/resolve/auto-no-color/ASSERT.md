## Expected

- `err` is nil.
- `resp.Enabled` is false: Auto + non-empty `noColorEnv` disables color on a TTY.

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
		t.Fatalf("Resolve(Auto, true, %q): unexpected error: %v", req.NoColorEnv, err)
	}
	if resp.Enabled {
		t.Fatal("Resolve(Auto, true, \"1\"): Enabled=true, want false (non-empty noColorEnv disables Auto)")
	}
}
```
