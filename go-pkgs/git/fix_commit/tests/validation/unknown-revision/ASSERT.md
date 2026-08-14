## Expected

- Exit 1. Stdout empty.
- Stderr is `Error: unknown revision: not-a-real-sha\n`.
- `HEAD` is unchanged.

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
	assertOutput(t, resp.Stderr, "Error: unknown revision: not-a-real-sha\n")
	head := runGitOutput(t, req.Dir, "rev-parse", "HEAD")
	if head == "" {
		t.Fatal("HEAD disappeared")
	}
}
```
