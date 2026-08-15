---
label: unix
explanation: chmod 000 directory fixture; default quiet permission skip
---

## Expected

- Exit code 0.
- Stdout contains `visible-repo` discovery line.
- Stderr is empty (no permission-skip warning).

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
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr without -v, got:\n%s", resp.Stderr)
	}
	roots := rootsFromArgs(req.Args)
	wantPath := absPath(t, filepath.Join(roots[0], "visible-repo"))
	if !strings.Contains(resp.Stdout, wantPath) {
		t.Fatalf("stdout should list visible-repo, got:\n%s", resp.Stdout)
	}
}
```