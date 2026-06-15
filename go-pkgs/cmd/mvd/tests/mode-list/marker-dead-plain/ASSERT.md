## Expected
- Exit code 0.
- 1 entry: moved (the latest) with `(dead)` marker.
- Root is NOT in the output (plain moves only show latest).

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

	moved := filepath.Join(req.WorkRoot, "repo-moved")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 picker entry, got %d:\n%s", len(lines), resp.Output)
	}

	parts := strings.SplitN(lines[0], " -> ", 2)
	if len(parts) != 2 || parts[1] != moved {
		t.Fatalf("expected full path %s, got:\n%s", moved, resp.Output)
	}

	if !strings.Contains(lines[0], "(dead)") {
		t.Fatalf("line should contain (dead) marker, got: %s", lines[0])
	}
	// Ensure (dead) is not part of a compound marker like (dead main)
	if strings.Contains(lines[0], "(dead main)") || strings.Contains(lines[0], "(dead worktree)") || strings.Contains(lines[0], "(dead external main)") {
		t.Fatalf("line should contain plain (dead), not a compound dead marker, got: %s", lines[0])
	}
}
```
