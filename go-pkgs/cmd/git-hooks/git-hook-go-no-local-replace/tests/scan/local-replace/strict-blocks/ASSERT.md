## Expected

- The command exits with a non-zero code.
- The `--strict` flag blocks even intra-repo replaces.
- The output line is `go.mod: => <abs>/sub`.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for strict mode with local replace, got 0\n%s", resp.Output)
	}
	got := strings.TrimSpace(resp.Output)
	if !strings.HasPrefix(got, "go.mod: => ") || !strings.HasSuffix(got, "/sub") {
		t.Fatalf("expected go.mod: => .../sub, got:\n%s", resp.Output)
	}
}
```