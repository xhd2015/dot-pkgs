## Expected

- Exit code 0.
- Stdout `intercepted\n`.
- Fake `--send` argument contains exactly one `\n` separating the two expanded array lines.
- First line starts with `wrk --no-interceptor`.
- Second line starts with `agent-run run`.
- No worktree under `{WRK_HOME}/worktrees/`.

## Exit Code

- 0

```go
import (
	"strings"
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
	args := assertInterceptorInvoked(t, req)
	var send string
	found := false
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--send" {
			send = args[i+1]
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("argv missing --send: %v", args)
	}
	if strings.Count(send, "\n") != 1 {
		t.Fatalf("send should contain exactly one newline between two lines, got %q", send)
	}
	lines := strings.Split(send, "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), send)
	}
	if !strings.HasPrefix(lines[0], "wrk --no-interceptor") {
		t.Fatalf("line0: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "agent-run run") {
		t.Fatalf("line1: %q", lines[1])
	}
	// second line should include shell_safe intent for /intent-route hello world
	wantSafe := shellSafeQuote("/intent-route " + req.TaskDesc)
	if !strings.Contains(lines[1], wantSafe) {
		t.Fatalf("line1 missing shell_safe intent %q: %q", wantSafe, lines[1])
	}
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
