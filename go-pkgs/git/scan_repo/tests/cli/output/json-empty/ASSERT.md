## Expected

- Exit code 0.
- Stdout is exactly `[]` (empty JSON array).
- Stderr is empty.

## Exit Code

- `0`.

```go
import (
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
	want := "[]"
	got := resp.Stdout
	if got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
```