## Expected

- The command exits with an error.
- The output indicates that patterns are required.

## Exit Code

- Exit code is non-zero.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for no patterns, got 0\n%s", resp.Output)
	}
	if resp.Output == "" {
		t.Fatalf("expected error output, got empty")
	}
}
```
