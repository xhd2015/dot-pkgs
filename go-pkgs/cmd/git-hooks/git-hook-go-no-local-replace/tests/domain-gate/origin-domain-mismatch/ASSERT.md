## Expected

- The command exits successfully with no output.
- The local replace is not detected because the domain gate skipped the hook.

## Exit Code

- Exit code is `0`.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for domain mismatch, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected no output for domain mismatch, got:\n%s", resp.Output)
	}
}

```
