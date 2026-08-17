## Expected

Only `old\r\n` becomes `new\r\n`; the indented closing-hash heading and LF neighbors are exact.

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
	assertExact(t, resp.Output, "lead\n   # Target ###\r\nnew\r\n# Next\nkeep\n")
	assertExact(t, resp.SecondOutput, "lead\n   # Target ###\r\nnew\r\n# Next\nkeep\n")
}
```

