## Expected
- Exit code 0.
- 1 picker entry: the latest (moved) path with `(aliases: dp)` marker.
- The original root path is NOT in the output.

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
	moved := filepath.Join(req.WorkRoot, "repo-moved")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 picker entry, got %d:\n%s", len(lines), resp.Output)
	}

	parts := strings.SplitN(lines[0], " -> ", 2)
	if len(parts) != 2 || parts[1] != moved {
		t.Fatalf("expected full path %s, got:\n%s", moved, resp.Output)
	}

	if !strings.Contains(lines[0], "(aliases: dp)") {
		t.Fatalf("expected alias marker (aliases: dp) on moved entry, got: %s", lines[0])
	}

	// Ensure root path does not appear as a standalone entry (it may
	// appear as a substring of the moved path, e.g. "repo" in "repo-moved")
	for _, part := range strings.Split(lines[0], " -> ") {
		if part == root {
			t.Fatalf("root path %s should NOT appear as a picker entry:\n%s", root, resp.Output)
		}
	}
}
```
