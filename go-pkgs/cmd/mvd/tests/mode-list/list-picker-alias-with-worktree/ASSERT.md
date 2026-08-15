## Expected
- Exit code 0.
- Output contains two entries: root (with alias annotation) and worktree (without alias annotation).
- The alias "myproj" appears on the root entry, not on the worktree entry.

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

	var rootLine string
	var wtLine string
	for _, line := range lines {
		parts := strings.SplitN(line, " -> ", 2)
		if len(parts) == 2 && parts[1] == root {
			rootLine = line
		}
		if len(parts) == 2 && parts[1] == wt {
			wtLine = line
		}
	}

	if rootLine == "" {
		t.Fatalf("root path %s not in picker output:\n%s", root, resp.Output)
	}
	if wtLine == "" {
		t.Fatalf("worktree path %s not in picker output:\n%s", wt, resp.Output)
	}

	if !strings.Contains(rootLine, "myproj") {
		t.Fatalf("alias 'myproj' should annotate root entry, got:\n%s", resp.Output)
	}
	if strings.Contains(wtLine, "myproj") {
		t.Fatalf("alias 'myproj' should NOT annotate worktree entry, got:\n%s", resp.Output)
	}
}
```
