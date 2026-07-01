## Expected

- The command exits with an error.
- The output contains the matched file name `main.go`.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for match, got 0\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "main.go") {
		t.Fatalf("expected output to contain matching file name, got:\n%s", resp.Output)
	}
}
```
