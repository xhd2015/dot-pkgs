## Expected

- The command exits with an error.
- The output contains the replace reference (e.g. `../another`).

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
	if !strings.Contains(resp.Output, "../another") {
		t.Fatalf("expected output to contain the local replace path, got:\n%s", resp.Output)
	}
}

```
