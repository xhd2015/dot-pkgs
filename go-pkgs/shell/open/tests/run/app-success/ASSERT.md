## Expected

- No error.
- Runner called once with `["open", "-a", "Visual Studio Code", <Target>]`.
- `Result.Args` is that argv.

## Side Effects

- One injected runner call; no application is launched.

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
	assertRunOnce(t, resp, []string{"open", "-a", "Visual Studio Code", req.Target})
	assertOut(t, resp, "launched")
}
```
