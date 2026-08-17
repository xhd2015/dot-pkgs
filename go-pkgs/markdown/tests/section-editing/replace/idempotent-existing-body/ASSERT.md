## Expected

Both `String` calls equal the original exactly.

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
	assertExact(t, resp.Output, "# Target\r\nsame\r\n## Child\nchild\r\n# Next\r\n")
	assertExact(t, resp.SecondOutput, "# Target\r\nsame\r\n## Child\nchild\r\n# Next\r\n")
}
```

