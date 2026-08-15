## Expected

- `NextTag` is `sub/v0.2.10`.

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
	if resp.NextTag != "sub/v0.2.10" {
		t.Fatalf("NextTag = %q, want sub/v0.2.10", resp.NextTag)
	}
}
```