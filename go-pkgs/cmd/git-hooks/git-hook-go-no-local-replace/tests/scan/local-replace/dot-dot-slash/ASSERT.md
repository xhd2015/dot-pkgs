## Expected

- The command exits with an error.
- The output line is `go.mod: => <abs>/another` (resolved sibling target dir).

## Exit Code

- Exit code is non-zero.

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for local replace, got 0\n%s", resp.Output)
	}
	wantSuffix := filepath.Join(filepath.Base(filepath.Dir(req.RepoDir)), "another")
	got := strings.TrimSpace(resp.Output)
	if !strings.HasPrefix(got, "go.mod: => ") || !strings.HasSuffix(got, "/another") {
		t.Fatalf("expected go.mod: => .../%s, got:\n%s", wantSuffix, resp.Output)
	}
}

```
