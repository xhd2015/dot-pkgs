## Expected

The error matches `ErrAmbiguousSection`; both duplicate sections remain exact.

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
	assertSentinel(t, err, markdown.ErrAmbiguousSection)
	assertExact(t, resp.Output, "# Same\none\n# Same\ntwo\n")
	assertExact(t, resp.SecondOutput, "# Same\none\n# Same\ntwo\n")
}
```

