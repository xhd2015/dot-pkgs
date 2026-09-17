## Expected

- No error.
- Resolve called once with `Home`.
- Runner called once with `[<CodePath>, "--goto", "<File>:7"]`.
- `Result.CodePath` mirrors the resolver.

## Side Effects

- One resolve call and one runner call.

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
	assertResolvedHome(t, resp, req.Home)
	assertDefaultRunner(t, resp, req, []string{req.CodePath, "--goto", req.File + ":7"})
}
```
