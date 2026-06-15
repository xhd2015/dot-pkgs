## Expected
- Exit code 0.
- 1 entry with no marker text (plain alive entry with no worktree, no alias).
- The display format is simply `$PATH -> $PATH` with no parenthesized marker.

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

	// The display part (before " -> ") should contain the path but no parenthesized marker
	display := parts[0]
	if strings.Contains(display, "(main)") || strings.Contains(display, "(worktree)") ||
		strings.Contains(display, "(dead") || strings.Contains(display, "(aliases") ||
		strings.Contains(display, "(external") {
		t.Fatalf("expected no marker on display, got: %s", display)
	}
}
```
