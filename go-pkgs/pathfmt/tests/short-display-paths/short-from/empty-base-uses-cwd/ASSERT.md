## Expected

- `resp.Display` is `"child"` (empty base uses cwd).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Display != "child" {
		t.Fatalf("empty base should use cwd: expected %q, got %q (cwd=%q path=%q)",
			"child", resp.Display, resp.Cwd, req.Path)
	}
}
```
