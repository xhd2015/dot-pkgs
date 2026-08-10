## Expected

- All five nodes of the chain appear in BFS order (PIDs 1..5).
- Proves `maxDepth=0` is treated as default 16 (not zero expansion).

## Errors

- `err` is nil.
- Truncated set (as if maxDepth were 0 meaning no children) is failure.

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
	want := linearChainFixture()
	assertProcsEqual(t, resp.Procs, want)
}
```
