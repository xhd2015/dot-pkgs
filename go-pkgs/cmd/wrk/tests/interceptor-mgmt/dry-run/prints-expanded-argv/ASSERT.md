## Expected Output

```
---
version: 2
---
echo
task=hi
```

## Expected

- Exit code 0.
- One expanded argv element per line with trailing newline after last line.
- Task flag `-t hi` feeds the `task` builtin → `task=hi`.

## Side Effects

- No interceptor binary exec (echo not required on PATH for dry-run print).
- No worktree under `{WRK_HOME}/worktrees/`.

## Exit Code

- 0

```go
import (
	"os"
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
echo
task=hi
`)
	// No worktrees created by management dry-run.
	wtRoot := filepath.Join(req.WrkHome, "worktrees")
	if entries, err := os.ReadDir(wtRoot); err == nil && len(entries) > 0 {
		t.Fatalf("dry-run must not create worktrees, got %v", entries)
	}
}
```
