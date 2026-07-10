## Expected

- Non-zero exit code.
- Stderr mentions the unknown variable (e.g. `no_such`) or template expansion error.
- Fake interceptor not invoked (expansion fails before exec).
- No worktree under `{WRK_HOME}/worktrees/`.

## Errors

- Unknown `${name}` / expansion error surfaces clearly; no silent native fallback.

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
		t.Fatalf("expected non-zero exit, got 0 stdout=%q", resp.Stdout)
	}
	combined := resp.Stderr + resp.Stdout
	if !strings.Contains(combined, "no_such") &&
		!strings.Contains(strings.ToLower(combined), "unknown") &&
		!strings.Contains(strings.ToLower(combined), "template") &&
		!strings.Contains(strings.ToLower(combined), "expand") {
		t.Fatalf("expected error mentioning unknown var/template, got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	assertInterceptorNotInvoked(t, req)
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
