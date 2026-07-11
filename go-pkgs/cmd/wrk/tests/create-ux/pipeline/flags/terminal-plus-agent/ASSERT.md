## Expected

- Exit 0.
- iTerm ForceNew at wt; follow-up contains agent-run + runner + prompt.
- Outer agent-run log empty.
- No space.

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
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertItermFollowUpHasAgentRun(t, script, req.TaskDesc)
	assertAgentRunNotInvoked(t, req)
}
```
