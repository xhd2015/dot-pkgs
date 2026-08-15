## Expected

- `resp.Formatted` is `"clean"`.

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
	if resp.Formatted != "clean" {
		t.Fatalf("formatted = %q, want clean", resp.Formatted)
	}
}
```