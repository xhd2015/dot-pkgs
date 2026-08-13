## Expected

- `err` is nil.
- `resp.Enabled` is true: empty `noColorEnv` does not disable Auto on a TTY.

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
	if !resp.Enabled {
		t.Fatal("Resolve(Auto, true, \"\"): Enabled=false, want true (empty noColorEnv does not disable)")
	}
}
```
