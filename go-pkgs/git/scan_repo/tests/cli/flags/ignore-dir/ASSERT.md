## Expected

- Exit code 0.
- Stdout is empty (`scratch` is ignored).
- Stderr is empty.

## Exit Code

- `0`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout with --ignore-dir scratch, got:\n%s", resp.Stdout)
	}
}
```