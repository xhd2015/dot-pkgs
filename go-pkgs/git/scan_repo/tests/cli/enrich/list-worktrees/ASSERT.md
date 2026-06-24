## Expected

- Exit code 0.
- Stdout has two path-sorted lines.
- `feature-a` line ends with `\tworktree`; `main` line ends with `\tmain`.

## Exit Code

- `0`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	roots := rootsFromArgs(req.Args)
	wantFeature := absPath(t, filepath.Join(roots[0], "feature-a")) + "\tworktree"
	wantMain := absPath(t, filepath.Join(roots[0], "main")) + "\tmain"
	lines := strings.Split(strings.TrimSuffix(resp.Stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), resp.Stdout)
	}
	if lines[0] != wantFeature {
		t.Fatalf("line[0] = %q, want %q", lines[0], wantFeature)
	}
	if lines[1] != wantMain {
		t.Fatalf("line[1] = %q, want %q", lines[1], wantMain)
	}
}
```