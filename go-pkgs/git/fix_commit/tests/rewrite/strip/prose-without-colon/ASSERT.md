## Expected

- Exit 1. Stdout empty.
- Same error as a missing trailer:
  `Error: commit message has no Co-authored-by line\n`.
- HEAD unchanged.

## Exit Code

- 1

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	requireHarnessOK(t, err)
	requireExit(t, resp, 1)
	if resp.Stdout != "" {
		t.Fatalf("stdout=%q want empty", resp.Stdout)
	}
	assertOutput(t, resp.Stderr, "Error: commit message has no Co-authored-by line\n")
	assertUnchangedSHA(t, req)
}
```
