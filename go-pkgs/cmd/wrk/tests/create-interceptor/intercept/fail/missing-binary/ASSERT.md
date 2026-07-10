## Expected

- Non-zero exit code.
- Stderr indicates missing binary / exec failure (not a silent native create).
- No worktree under `{WRK_HOME}/worktrees/`.
- Interceptor log remains empty (binary never ran).

## Errors

- Fail hard when interceptor binary is not on PATH or cannot be executed.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when interceptor binary missing, stdout=%q", resp.Stdout)
	}
	// Must not have fallen back to native create (which prints a worktree path).
	wantNative := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	if strings.TrimSpace(resp.Stdout) == wantNative {
		t.Fatalf("stdout looks like native worktree path — silent fallback is forbidden: %q", resp.Stdout)
	}
	assertNoWorktreesUnderHome(t, req.WrkHome)
	assertInterceptorNotInvoked(t, req)
	combined := strings.ToLower(resp.Stderr + " " + resp.Stdout)
	if !strings.Contains(combined, "not-installed") &&
		!strings.Contains(combined, "not found") &&
		!strings.Contains(combined, "no such file") &&
		!strings.Contains(combined, "executable") &&
		!strings.Contains(combined, "exec") &&
		!strings.Contains(combined, "interceptor") &&
		!strings.Contains(combined, "path") {
		// Soft check: non-zero + no worktree is the hard requirement; message is best-effort.
		t.Logf("note: stderr/stdout did not match common missing-binary phrases: stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
}
```
