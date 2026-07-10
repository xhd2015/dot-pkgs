## Expected

- Exit code 0.
- Stdout `intercepted\n`.
- Fake argv includes `--send` followed by a single arg containing two lines (joined with `\n`):
  1. `wrk --no-interceptor …` with shell-safe-quoted create args
  2. `agent-run run …` ending with shell-safe-encoded intent token for `/intent-route fix "quoted" task`
- The shell_safe token is a single POSIX single-quoted word; embedded `"` characters appear inside the single-quoted string and do not terminate the word.
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
	// Find --send value
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
	lines := strings.Split(send, "\n")
	if len(lines) != 2 {
		t.Fatalf("send should be 2 lines joined by \\n, got %d: %q", len(lines), send)
	}
	if !strings.HasPrefix(lines[0], "wrk --no-interceptor") {
		t.Fatalf("line0 should start with wrk --no-interceptor, got %q", lines[0])
	}
	// args_shell_safe should include shell-safe -t and task text
	wantTaskQ := shellSafeQuote(req.TaskDesc)
	wantFlagQ := shellSafeQuote("-t")
	if !strings.Contains(lines[0], wantFlagQ) || !strings.Contains(lines[0], wantTaskQ) {
		t.Fatalf("line0 should embed args_shell_safe with %s %s; got %q", wantFlagQ, wantTaskQ, lines[0])
	}
	rawIntent := "/intent-route " + req.TaskDesc
	wantSafe := shellSafeQuote(rawIntent)
	if !strings.Contains(lines[1], wantSafe) {
		t.Fatalf("line1 should contain shell_safe intent %q; got %q", wantSafe, lines[1])
	}
	if !strings.Contains(lines[1], `"`) {
		t.Fatalf("line1 should still carry double-quote characters inside the shell_safe word: %q", lines[1])
	}
	assertNoWorktreesUnderHome(t, req.WrkHome)
}
```
