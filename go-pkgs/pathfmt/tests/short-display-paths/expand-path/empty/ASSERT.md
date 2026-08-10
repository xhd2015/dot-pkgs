## Expected

- `resp.Display` is `""` (unchanged).

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
	if resp.Display != "" {
		t.Fatalf("expected empty display for empty input, got %q", resp.Display)
	}
}```
