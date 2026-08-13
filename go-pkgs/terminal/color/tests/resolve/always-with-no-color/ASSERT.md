## Expected

- `err` is nil.
- `resp.Enabled` is true: Always ignores `noColorEnv` (flags win).

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
		t.Fatalf("Resolve(Always, true, %q): unexpected error: %v", req.NoColorEnv, err)
	}
	if !resp.Enabled {
		t.Fatal("Resolve(Always, true, \"1\"): Enabled=false, want true (flags win over NO_COLOR)")
	}
}
```
