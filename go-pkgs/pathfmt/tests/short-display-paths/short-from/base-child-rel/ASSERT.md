## Expected

- `resp.Display` is `"child"`.

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
		t.Fatalf("expected %q, got %q (base=%q path=%q)", "child", resp.Display, req.BaseDir, req.Path)
	}
}
```
