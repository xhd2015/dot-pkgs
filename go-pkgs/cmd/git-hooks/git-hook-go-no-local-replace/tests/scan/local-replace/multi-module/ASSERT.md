## Expected

- The command succeeds because the subdirectory replace points inside the same git repo.
- The command emits no output.

## Exit Code

- Exit code is 0.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for multi-module intra-repo replace, got %d\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected no output for allowed intra-repo replace, got:\n%s", resp.Output)
	}
}

```
