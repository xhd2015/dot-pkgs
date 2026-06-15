## Expected
- Exit code 0.
- 2 entries: root with `(main)`, worktree with `(dead worktree)`.
- Both paths appear even though worktree does not exist on disk.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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

	var rootLine, wtLine string
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

	if !strings.Contains(rootLine, "(main)") {
		t.Fatalf("root line should contain (main), got: %s", rootLine)
	}
	if !strings.Contains(wtLine, "(dead worktree)") {
		t.Fatalf("wt line should contain (dead worktree), got: %s", wtLine)
	}
}
```
