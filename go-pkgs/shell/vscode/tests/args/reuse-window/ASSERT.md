## Expected

- No error.
- `Args` = `[<CodePath>, "-r", <Dir>]`.
- `-r` comes before the directory, which is the last element.

## Side Effects

- None: `Args` is pure.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertArgs(t, resp.Args, []string{req.CodePath, "-r", req.Dir})
}
```
