## Expected

- The command exits with an error.
- The output states that `--unknown` is an unknown flag.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code for unknown flag\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "unknown flag: --unknown") {
		t.Fatalf("expected unknown flag error, got:\n%s", resp.Output)
	}
}
```
