## Expected

- `err != nil`.
- No argv is produced.

## Side Effects

- None: `FileArgs` is pure.

## Errors

- `vscode: file is empty`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	if len(resp.Args) != 0 {
		t.Fatalf("Args = %#v, want none", resp.Args)
	}
}
```
