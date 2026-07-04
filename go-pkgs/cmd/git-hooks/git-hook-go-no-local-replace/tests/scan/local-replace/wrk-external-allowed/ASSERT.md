## Expected

- The command exits with code 0 (replace target is inside the scanning worktree).
- No output on stdout.

## Exit Code

- Exit code is 0.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for replace under worktree/external/, got %d\n%s", resp.ExitCode, resp.Output)
	}
	if strings.TrimSpace(resp.Output) != "" {
		t.Fatalf("expected no output, got %q", resp.Output)
	}
}

```