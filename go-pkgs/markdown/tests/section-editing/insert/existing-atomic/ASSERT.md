## Expected

The error matches `ErrSectionExists`; the existing section is unchanged.

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
	assertSentinel(t, err, markdown.ErrSectionExists)
	assertExact(t, resp.Output, "# Added\nbody\n# User\n")
	assertExact(t, resp.SecondOutput, "# Added\nbody\n# User\n")
}
```

