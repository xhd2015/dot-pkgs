## Expected

- No error.
- `Args` = `[<CodePath>, "--goto", "<File>:42"]` — one `file:line` token.

## Side Effects

- None: `FileArgs` is pure.

## Errors

- None for a non-empty file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertArgs(t, resp.Args, []string{req.CodePath, "--goto", req.File + ":42"})
}
```
