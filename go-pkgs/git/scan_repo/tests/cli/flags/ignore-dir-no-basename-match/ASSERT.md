## Expected

- Exit code 0.
- Stdout lists the hidden repo — `scratch` as a relative `--ignore-dir` value does not skip by basename.
- Stderr is empty.

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
	wantPath := absPath(t, filepath.Join(roots[0], "scratch", "hidden-repo"))
	wantLine := wantPath + "\tmain"
	got := strings.TrimSuffix(resp.Stdout, "\n")
	if got != wantLine {
		t.Fatalf("stdout = %q, want %q (repo should not be skipped)", got, wantLine)
	}
}
```