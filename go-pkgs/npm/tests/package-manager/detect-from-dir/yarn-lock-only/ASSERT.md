## Expected

- `resp.Trace.Manager` is `yarn`.

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
	if resp.Trace.Manager != npm.ManagerYarn {
		t.Fatalf("manager = %q, want yarn", resp.Trace.Manager)
	}
}```
