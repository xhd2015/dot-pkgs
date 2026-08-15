## Expected

- `resp.Trace.Manager` is `pnpm`.
- `resp.Trace.HasPackageJSON` is true.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Trace.Manager != npm.ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", resp.Trace.Manager)
	}
	if !resp.Trace.HasPackageJSON {
		t.Fatal("expected HasPackageJSON = true")
	}
}```
