## Expected

The error matches `ErrInvalidHeader`; output remains empty.

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
	assertSentinel(t, err, markdown.ErrInvalidHeader)
	assertExact(t, resp.Output, "")
	assertExact(t, resp.SecondOutput, "")
}
```

