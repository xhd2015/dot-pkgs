## Expected
- Exit code 0.
- 1 entry: root with `(aliases: myproj)` marker (no worktree, so no main marker to combine with).

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

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 picker entry, got %d:\n%s", len(lines), resp.Output)
	}

	parts := strings.SplitN(lines[0], " -> ", 2)
	if len(parts) != 2 || parts[1] != root {
		t.Fatalf("expected full path %s, got:\n%s", root, resp.Output)
	}

	if !strings.Contains(lines[0], "(aliases: myproj)") {
		t.Fatalf("line should contain (aliases: myproj), got: %s", lines[0])
	}
	if !strings.Contains(lines[0], "myproj") {
		t.Fatalf("line should contain alias 'myproj', got: %s", lines[0])
	}
}
```
