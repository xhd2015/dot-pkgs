## Expected

- Non-zero exit code.
- Stdout is empty.
- Stderr reports the unknown flag.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for unknown flag\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	combined := resp.Stderr
	if !strings.Contains(combined, "unknown") && !strings.Contains(combined, "--unknown") {
		t.Fatalf("expected unknown flag error, got stderr:\n%s", resp.Stderr)
	}
}
```