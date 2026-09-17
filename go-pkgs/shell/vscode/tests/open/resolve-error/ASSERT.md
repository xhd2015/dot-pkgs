## Expected

- `err != nil` and it carries the resolver's message.
- Resolution was attempted once, with `Home`.
- The runner is not called.

## Side Effects

- One injected resolve call; no process launch.

## Errors

- The resolver error, wrapped with the `vscode:` prefix.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	if !strings.Contains(err.Error(), "not found") || !strings.HasPrefix(err.Error(), "vscode:") {
		t.Fatalf("err = %v, want the resolver error wrapped with a vscode: prefix", err)
	}
	assertResolvedHome(t, resp, req.Home)
	assertNoRun(t, resp)
}
```
