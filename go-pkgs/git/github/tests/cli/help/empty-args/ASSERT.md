## Expected

- `resp.ExitCode` is 0 (not an error path).
- `resp.Stdout` is top-level usage: non-empty, ends with `\n`, mentions `repo`.
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation.

## Errors

- `err` from harness is nil.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHelpStdout(t, resp, "repo")
}
```
