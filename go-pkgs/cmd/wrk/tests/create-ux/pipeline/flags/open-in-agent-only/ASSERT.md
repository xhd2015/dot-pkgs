## Expected

- Exit 0; path printed (worktree includes task slug).
- Outer agent-run invoked with cwd=worktree and default argv/prompt.
- No space; no iterm.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	assertAgentRunInvoked(t, req, wt, req.TaskDesc)
}
```
