## Expected

- `resp.Trace.Manager` is `pnpm`.
- Multiple signals were collected.

## Errors

- `err` is nil.

```go
import (
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Trace.Manager != npm.ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", resp.Trace.Manager)
	}
	if len(resp.Trace.Signals) < 4 {
		t.Fatalf("expected multiple signals, got %d", len(resp.Trace.Signals))
	}
}```
