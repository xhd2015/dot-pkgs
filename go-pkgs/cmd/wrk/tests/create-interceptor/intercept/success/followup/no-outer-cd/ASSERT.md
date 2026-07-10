## Expected

- Exit code 0.
- Stdout is fake output `intercepted\n` (not a worktree path).
- Follow-up file exists but contains **no** `cd` line from outer wrk.
- Fake was invoked.
- No worktree under `{WRK_HOME}/worktrees/`.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
intercepted
`)
	assertInterceptorInvoked(t, req)
	assertFollowupHasNoCD(t, req)
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
