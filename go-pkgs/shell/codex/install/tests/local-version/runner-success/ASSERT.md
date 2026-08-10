## Expected

- `err == nil`.
- `resp.Version == "codex-cli 0.147.0"` (raw command output).
- Injected `RunVersion` was called at least once.

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
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Version", resp.Version, "codex-cli 0.147.0")
	if len(resp.RunVersionCalls) < 1 {
		t.Fatalf("RunVersionCalls = %#v, want >= 1", resp.RunVersionCalls)
	}
}
```
