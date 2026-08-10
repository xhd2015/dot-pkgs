## Expected

- BFS includes root at depth 0 and child at depth 1 only:
  - `{1, 0, "root"}`, `{2, 1, "child"}`
- Grandchild PID 3 is **not** included when `maxDepth=1`.

## Errors

- `err` is nil.
- Including PID 3 or omitting root is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	want := []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "root"},
		{PID: 2, PPID: 1, Cmd: "child"},
	}
	assertProcsEqual(t, resp.Procs, want)
}
```
