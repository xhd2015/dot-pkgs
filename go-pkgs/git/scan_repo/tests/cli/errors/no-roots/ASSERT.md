## Expected

- Non-zero exit code.
- Stdout is empty.
- Stderr mentions that roots are required.

## Errors

- `RunCLI` returns a non-nil error (mapped to exit code 1 by harness).

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for missing --root\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	msg := strings.ToLower(resp.Stderr)
	if !strings.Contains(msg, "root") {
		t.Fatalf("stderr should mention root, got:\n%s", resp.Stderr)
	}
}
```