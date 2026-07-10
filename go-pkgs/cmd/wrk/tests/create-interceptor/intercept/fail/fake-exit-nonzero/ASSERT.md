## Expected

- Exit code 3 (propagated from interceptor child).
- Fake was invoked.
- No worktree under `{WRK_HOME}/worktrees/`.

## Exit Code

- 3

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 3 {
		t.Fatalf("expected exit 3, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertInterceptorInvoked(t, req)
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
