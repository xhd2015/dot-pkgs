## Expected

- `err != nil`.
- Neither the resolver nor the runner is called.

## Side Effects

- No resolution work and no process launch.

## Errors

- `vscode: dir is empty`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	assertNoResolve(t, resp)
	assertNoRun(t, resp)
}
```
