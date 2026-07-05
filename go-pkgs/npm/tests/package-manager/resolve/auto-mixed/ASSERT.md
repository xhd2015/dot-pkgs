## Expected

- `resp.Manager` is `pnpm`.

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
	if resp.Manager != npm.ManagerPnpm {
		t.Fatalf("manager = %q, want pnpm", resp.Manager)
	}
}```
