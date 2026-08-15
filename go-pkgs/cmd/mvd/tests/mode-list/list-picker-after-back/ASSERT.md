## Expected
- Exit code 0.
- Output contains exactly one picker entry: the root path (worktree is gone after --back).

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	root := filepath.Join(req.WorkRoot, "repo")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 picker entry after back, got %d:\n%s", len(lines), resp.Output)
	}

	parts := strings.SplitN(lines[0], " -> ", 2)
	if len(parts) != 2 || parts[1] != root {
		t.Fatalf("expected full path %s, got:\n%s", root, resp.Output)
	}
}
```
