## Expected

- No error.
- Runner called exactly once with `["open", <Target>]`.
- `Result.Args` is that argv and `Result.Out` is the runner's output.

## Side Effects

- One injected runner call; no real process.

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
	assertRunOnce(t, resp, []string{"open", req.Target})
	assertArgs(t, resp.Args, []string{"open", req.Target})
	assertOut(t, resp, "launched")
}
```
