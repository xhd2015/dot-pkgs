## Expected

The selected H1 and child H2 disappear without changing either neighbor.

## Errors

- The returned error and sentinel identity must match this scenario.
- Exact text comparisons include every byte and line ending.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/markdown"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	assertExact(t, resp.Output, "lead\n# A\nx\n# C\nz\n")
	assertExact(t, resp.SecondOutput, "lead\n# A\nx\n# C\nz\n")
}
```

