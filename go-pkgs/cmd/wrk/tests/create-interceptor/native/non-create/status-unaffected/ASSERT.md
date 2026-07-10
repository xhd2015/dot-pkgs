## Expected

- Exit code 0.
- Stdout is a status block (contains `Dir:` for the root checkout).
- Fake interceptor not invoked.
- No worktrees created under `{WRK_HOME}/worktrees/`.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Dir:") {
		t.Fatalf("expected status stdout with Dir:, got %q", resp.Stdout)
	}
	assertInterceptorNotInvoked(t, req)
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
