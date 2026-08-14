## Expected

- Exit 1. Stdout empty.
- Stderr is `Error: nothing to change\n`.
- `HEAD` / `master` still the old SHA.

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
	assertOutput(t, resp.Stderr, "Error: nothing to change\n")
	assertUnchangedSHA(t, req)
	if got := runGitOutput(t, req.Dir, "rev-parse", "refs/heads/master"); got != req.OldSHA {
		t.Fatalf("master moved to %s", got)
	}
}
```
