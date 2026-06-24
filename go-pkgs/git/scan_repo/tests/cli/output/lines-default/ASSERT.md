## Expected

- Exit code 0.
- Stdout has two tab-separated lines sorted by path: `alpha` then `zebra`.
- Each line format: `{path}\tmain`.

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
	wantAlpha := absPath(t, filepath.Join(roots[0], "alpha")) + "\tmain"
	wantZebra := absPath(t, filepath.Join(roots[0], "zebra")) + "\tmain"
	lines := strings.Split(strings.TrimSuffix(resp.Stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), resp.Stdout)
	}
	if lines[0] != wantAlpha {
		t.Fatalf("line[0] = %q, want %q", lines[0], wantAlpha)
	}
	if lines[1] != wantZebra {
		t.Fatalf("line[1] = %q, want %q", lines[1], wantZebra)
	}
}
```