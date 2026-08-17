## Expected

The inserted heading and content use CRLF while existing LF bytes remain LF.

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
	assertExact(t, resp.Output, "lead\r\n# Added\r\nnew\r\n# User\nbody\n")
	assertExact(t, resp.SecondOutput, "lead\r\n# Added\r\nnew\r\n# User\nbody\n")
}
```

