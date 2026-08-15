## Expected

- Exit code 0.
- Stdout has two lines (repo-a before repo-b by path).
- Each line ends with `\tmain`.

## Exit Code

- `0`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	roots := rootsFromArgs(req.Args)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots in args, got %v", roots)
	}
	wantA := absPath(t, filepath.Join(roots[0], "repo-a")) + "\tmain"
	wantB := absPath(t, filepath.Join(roots[1], "repo-b")) + "\tmain"
	lines := strings.Split(strings.TrimSuffix(resp.Stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), resp.Stdout)
	}
	if lines[0] != wantA {
		t.Fatalf("line[0] = %q, want %q", lines[0], wantA)
	}
	if lines[1] != wantB {
		t.Fatalf("line[1] = %q, want %q", lines[1], wantB)
	}
}
```