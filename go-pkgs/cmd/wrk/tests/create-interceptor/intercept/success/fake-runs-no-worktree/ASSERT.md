## Expected Output

```text
intercepted
```

## Expected

- Exit code 0.
- Stdout is exactly `intercepted\n` (fake tool output; trailing newline).
- Fake invoked with argv starting with `kool`, `space`, `create`, `--work-dir`, then resolved work_dir abs path (`MainRepo`).
- No entries under `{WRK_HOME}/worktrees/`.
- Stderr empty (or only implementation-defined noise; prefer empty).

## Side Effects

- Outer wrk did not run native `git worktree add`.
- Outer `events.jsonl` last command is still `create` with exit 0.

## Exit Code

- 0

```go
import (
	"path/filepath"
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
	if len(args) < 5 {
		t.Fatalf("expected >=5 argv elements, got %v", args)
	}
	if args[0] != fakeInterceptorName || args[1] != "space" || args[2] != "create" || args[3] != "--work-dir" {
		t.Fatalf("unexpected argv prefix: %v", args)
	}
	wantDir, err := filepath.EvalSymlinks(req.MainRepo)
	if err != nil {
		wantDir = req.MainRepo
	}
	gotDir, err := filepath.EvalSymlinks(args[4])
	if err != nil {
		gotDir = args[4]
	}
	if gotDir != wantDir {
		t.Fatalf("work_dir: want %q, got %q (raw %q)", wantDir, gotDir, args[4])
	}
	assertNoWorktreesUnderHome(t, req.WrkHome)
	assertLastEventCommand(t, req.WrkHome, "create", 0)
}
```
