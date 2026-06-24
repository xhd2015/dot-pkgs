## Expected

- Exit code 0.
- Stderr is empty.
- Stdout is exactly one line: `{abs-path}\tmain`.

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
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	roots := rootsFromArgs(req.Args)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root in args, got %v", roots)
	}
	wantPath := absPath(t, filepath.Join(roots[0], "my-repo"))
	wantLine := wantPath + "\tmain"
	got := strings.TrimSuffix(resp.Stdout, "\n")
	if got != wantLine {
		t.Fatalf("stdout = %q, want %q", got, wantLine)
	}
}
```