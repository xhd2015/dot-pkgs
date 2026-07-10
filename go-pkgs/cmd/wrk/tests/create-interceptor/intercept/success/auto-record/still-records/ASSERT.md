## Expected

- Exit code 0.
- Fake ran; no worktree under WRK_HOME.
- `projects.json` contains the main repo with `source: "auto"`.
- Last `events.jsonl` entry has `command: "create"` and `exit_code: 0`.

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
	assertNoWorktreesUnderHome(t, req.WrkHome)
	assertProjectAutoRecorded(t, req.WrkHome, req.MainRepo)
	assertLastEventCommand(t, req.WrkHome, "create", 0)
}
```
