## Expected

- Exit code 0.
- Stdout is one line: `{path}\tmain\torigin:xhd2015/lifelog@github.com`.

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
	wantPath := absPath(t, filepath.Join(roots[0], "proj"))
	wantLine := wantPath + "\tmain\torigin:xhd2015/lifelog@github.com"
	got := strings.TrimSuffix(resp.Stdout, "\n")
	if got != wantLine {
		t.Fatalf("stdout = %q, want %q", got, wantLine)
	}
}
```