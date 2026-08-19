## Expected

- `resp.OK` is false (the space is required and is not in the haystack).

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
	if resp.OK {
		t.Fatalf("expected Match(%q, %q) !OK (space is literal)", req.Haystack, req.Query)
	}
}
```
