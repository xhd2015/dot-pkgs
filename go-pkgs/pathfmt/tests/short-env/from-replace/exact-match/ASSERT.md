## Expected

- `resp.Display` is exactly `"$X"`.

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
	if resp.Display != "$X" {
		t.Fatalf("expected %q, got %q (path=%q)", "$X", resp.Display, req.Path)
	}
}
```
