## Expected

- The command exits with a non-zero code.
- The `--strict` flag blocks intra-repo replaces even in subdirectories.
- The output line is `sub/go.mod: => <abs>/sub/local`.

## Exit Code

- Exit code is non-zero.

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for strict mode with intra-repo replace in sub, got 0\n%s", resp.Output)
	}
	got := strings.TrimSpace(resp.Output)
	if got != "sub/go.mod: => "+filepath.Join(req.RepoDir, "sub", "local") &&
		!(strings.HasPrefix(got, "sub/go.mod: => ") && strings.HasSuffix(got, "/sub/local")) {
		t.Fatalf("expected sub/go.mod: => .../sub/local, got:\n%s", resp.Output)
	}
}
```