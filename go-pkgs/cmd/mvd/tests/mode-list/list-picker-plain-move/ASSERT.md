## Expected
- Exit code 0.
- Output contains exactly one picker entry: the latest (moved) path.
- The original root path is NOT in the output (regression: plain moves should not show old locations).

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
	moved := filepath.Join(req.WorkRoot, "repo-moved")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 picker entry for plain move, got %d:\n%s", len(lines), resp.Output)
	}

	parts := strings.SplitN(lines[0], " -> ", 2)
	if len(parts) != 2 || parts[1] != moved {
		t.Fatalf("expected full path %s, got:\n%s", moved, resp.Output)
	}

	if strings.Contains(parts[1]+" ", root+" ") || parts[1] == root {
		// extra safety: full path must not be root
		t.Fatalf("root path %s should NOT be the picker entry for plain move:\n%s", root, resp.Output)
	}
}
```
