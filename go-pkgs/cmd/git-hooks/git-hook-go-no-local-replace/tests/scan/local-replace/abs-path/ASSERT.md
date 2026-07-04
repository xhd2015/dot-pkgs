## Expected

- The command exits with an error.
- The output line is `go.mod: => /tmp/somepkg`.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for local replace, got 0\n%s", resp.Output)
	}
	got := strings.TrimSpace(resp.Output)
	if !strings.HasPrefix(got, "go.mod: => ") || !strings.HasSuffix(got, "/tmp/somepkg") {
		t.Fatalf("expected go.mod: => .../tmp/somepkg, got:\n%s", resp.Output)
	}
}

```
