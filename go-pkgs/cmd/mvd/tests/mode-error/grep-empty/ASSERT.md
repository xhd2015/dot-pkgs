## Expected

- Non-zero exit code.
- Combined output mentions that `--grep` requires a non-empty pattern/filter.

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for empty --grep, got 0: %s", resp.Output)
	}
	lower := strings.ToLower(resp.Output)
	if !strings.Contains(lower, "grep") {
		t.Fatalf("expected error mentioning --grep, got: %s", resp.Output)
	}
	if !strings.Contains(lower, "non-empty") && !strings.Contains(lower, "empty") && !strings.Contains(lower, "require") {
		t.Fatalf("expected empty-pattern error for --grep, got: %s", resp.Output)
	}
}
```
