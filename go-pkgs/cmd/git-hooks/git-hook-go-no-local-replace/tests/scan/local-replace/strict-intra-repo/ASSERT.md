## Expected

- The command exits with a non-zero code.
- The `--strict` flag blocks intra-repo replaces even in subdirectories.
- The output references the subdirectory's local replace.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for strict mode with intra-repo replace in sub, got 0\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "./local") {
		t.Fatalf("expected output to contain the local replace path, got:\n%s", resp.Output)
	}
}
```