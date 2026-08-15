## Expected
- Exit code 0.
- Output contains two picker entries: one for the root repo, one for the worktree.
- Both paths appear in the output.

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
	wt := filepath.Join(req.WorkRoot, "feature")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 picker entries, got %d:\n%s", len(lines), resp.Output)
	}

	foundRoot := false
	foundWt := false
	for _, line := range lines {
		parts := strings.SplitN(line, " -> ", 2)
		if len(parts) == 2 && parts[1] == root {
			foundRoot = true
		}
		if len(parts) == 2 && parts[1] == wt {
			foundWt = true
		}
	}
	if !foundRoot {
		t.Fatalf("root path %s not in picker output:\n%s", root, resp.Output)
	}
	if !foundWt {
		t.Fatalf("worktree path %s not in picker output:\n%s", wt, resp.Output)
	}
}
```
