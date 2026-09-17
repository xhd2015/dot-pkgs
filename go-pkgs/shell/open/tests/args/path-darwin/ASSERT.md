## Expected

- No error.
- `Args` = `["open", <Target>]`.

## Side Effects

- None: `Args` is pure.

## Errors

- None for a valid target on a supported platform.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertArgs(t, resp.Args, []string{"open", req.Target})
}
```
